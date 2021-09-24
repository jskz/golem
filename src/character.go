/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const UnauthenticatedUsername = "unnamed"

type Job struct {
	Id                         uint    `json:"id"`
	Name                       string  `json:"name"`
	DisplayName                string  `json:"display_name"`
	Playable                   bool    `json:"playable"`
	ExperienceRequiredModifier float64 `json:"experience_required_modifier"`
}

type Race struct {
	Id          uint   `json:"id"`
	Name        string `json:"race"`
	DisplayName string `json:"display_name"`
	Playable    bool   `json:"playable"`
}

const LevelAdmin = 60
const LevelHero = 50

/* These flag constants are shared by both PCs and NPCs */
const (
	CHAR_IS_PLAYER  = 1
	CHAR_SENTINEL   = 1 << 1
	CHAR_STAY_AREA  = 1 << 2
	CHAR_AGGRESSIVE = 1 << 3
)

const (
	RESIST_BASH   = 1
	RESIST_FIRE   = 1 << 1
	RESIST_COLD   = 1 << 2
	RESIST_SHOCK  = 1 << 3
	RESIST_POISON = 1 << 4
)

const (
	IMMUNE_BASH   = 1
	IMMUNE_FIRE   = 1 << 1
	IMMUNE_COLD   = 1 << 2
	IMMUNE_SHOCK  = 1 << 3
	IMMUNE_POISON = 1 << 4
)

const (
	SUSCEPT_BASH   = 1
	SUSCEPT_FIRE   = 1 << 1
	SUSCEPT_COLD   = 1 << 2
	SUSCEPT_SHOCK  = 1 << 3
	SUSCEPT_POISON = 1 << 4
)

const (
	PositionDead     = 0
	PositionStunned  = 1
	PositionSleeping = 2
	PositionResting  = 3
	PositionSitting  = 4
	PositionFighting = 5
	PositionStanding = 8
)

/*
 * This character structure is shared by both player-characters (human beings
 * connected through a session instance available via the client pointer.)
 */
type Character struct {
	game   *Game
	client *Client

	inventory *LinkedList

	equipment []*ObjectInstance

	pages      [][]byte
	pageSize   int
	pageCursor int

	Room      *Room           `json:"room"`
	Combat    *Combat         `json:"combat"`
	Fighting  *Character      `json:"fighting"`
	casting   *CastingContext `json:"casting"`
	following *Character

	id int `json:"id"`

	name             string `json:"name"`
	shortDescription string `json:"shortDescription"`
	longDescription  string `json:"longDescription"`
	description      string `json:"description"`

	wizard     bool
	job        *Job
	race       *Race
	level      uint
	experience uint
	practices  int

	skills map[uint]*Proficiency

	flags int
	afk   *AwayFromKeyboard

	health     int
	maxHealth  int
	mana       int
	maxMana    int
	stamina    int
	maxStamina int

	position int

	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Constitution int `json:"constitution"`
	Charisma     int `json:"charisma"`
	Luck         int `json:"luck"`

	temporaryHash string
}

func (ch *Character) experienceRequiredForLevel(level int) int {
	required := int(500*(level*level) - (500 * level))

	if ch.job != nil {
		required = int(float64(required) * ch.job.ExperienceRequiredModifier)
	}

	return required
}

func (game *Game) AttemptLogin(username string, password string) bool {
	var hash string

	row := game.db.QueryRow(`
		SELECT
			password_hash
		FROM
			player_characters
		WHERE
			username = ?
		AND
			deleted_at IS NULL
	`, username)

	err := row.Scan(&hash)
	if err != nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func (ch *Character) onZoneUpdate() {
}

func (ch *Character) onUpdate() {
	/* Regen, script hooks, etc. */
	if ch.health < ch.maxHealth {
		ch.health = int(math.Min(float64(ch.maxHealth), float64(ch.health+10)))
	}

	if ch.stamina < ch.maxStamina {
		ch.stamina = int(math.Min(float64(ch.maxStamina), float64(ch.stamina+15)))
	}
}

func (ch *Character) Finalize() error {
	if ch.client == nil || ch.game == nil {
		/* If somehow an NPC were to try to save, do not allow it. */
		return nil
	}

	result, err := ch.game.db.Exec(`
		INSERT INTO
			player_characters(username, password_hash, wizard, room_id, race_id, job_id, level, experience, practices, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ch.name, ch.temporaryHash, 0, RoomLimbo, ch.race.Id, ch.job.Id, ch.level, ch.experience, ch.practices, ch.health, ch.maxHealth, ch.mana, ch.maxMana, ch.stamina, ch.maxStamina, ch.Strength, ch.Dexterity, ch.Intelligence, ch.Wisdom, ch.Constitution, ch.Charisma, ch.Luck)
	ch.temporaryHash = ""
	if err != nil {
		log.Printf("Failed to finalize new character: %v.\r\n", err)
		return err
	}

	userId, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve insert id: %v.\r\n", err)
		return err
	}

	ch.id = int(userId)

	limbo, err := ch.game.LoadRoomIndex(RoomLimbo)
	if err != nil {
		return err
	}

	ch.Room = limbo
	return nil
}

func (ch *Character) Save() bool {
	if ch.client == nil || ch.game == nil {
		/* If somehow an NPC were to try to save, do not allow it. */
		return false
	}

	var roomId uint = RoomLimbo
	if ch.Room != nil {
		roomId = ch.Room.Id
	}
	result, err := ch.game.db.Exec(`
		UPDATE
			player_characters
		SET
			wizard = ?,
			room_id = ?,
			race_id = ?,
			job_id = ?,
			level = ?,
			experience = ?,
			practices = ?,
			health = ?,
			max_health = ?,
			mana = ?,
			max_mana = ?,
			stamina = ?,
			max_stamina = ?,
			stat_str = ?,
			stat_dex = ?,
			stat_int = ?,
			stat_wis = ?,
			stat_con = ?,
			stat_cha = ?,
			stat_lck = ?,
			updated_at = NOW()
		WHERE
			id = ?
	`, ch.wizard, roomId, ch.race.Id, ch.job.Id, ch.level, ch.experience, ch.practices, ch.health, ch.maxHealth, ch.mana, ch.maxMana, ch.stamina, ch.maxStamina, ch.Strength, ch.Dexterity, ch.Intelligence, ch.Wisdom, ch.Constitution, ch.Charisma, ch.Luck, ch.id)
	if err != nil {
		log.Printf("Failed to save character: %v.\r\n", err)
		return false
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Failed to retrieve number of rows affected: %v.\r\n", err)
		return false
	}

	return rowsAffected == 1
}

func (ch *Character) attachObject(obj *ObjectInstance) error {
	obj.reify()

	return nil
}

func (ch *Character) detachObject(obj *ObjectInstance) error {
	result, err := ch.game.db.Exec(`
		DELETE FROM
			player_character_object
		WHERE
			player_character_id = ?
		AND
			object_instance_id = ?`, ch.id, obj.id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected != 1 {
		/* Weird, but not fatal */
		return nil
	}

	return nil
}

func (game *Game) SavePlayerInventory(ch *Character) error {
	/* Object instances whose records have dirtied */
	var updating []*ObjectInstance = make([]*ObjectInstance, 0)

	/* Iterate over all objects in this player's inventory */
	for iter := ch.inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		/* If this is a container, ensure that all contained object instances are also updated */
		if obj.contents != nil && obj.contents.Count > 0 {
			for containerIter := obj.contents.Head; containerIter != nil; containerIter = containerIter.Next {
				containedObj := containerIter.Value.(*ObjectInstance)

				updating = append(updating, containedObj)
			}
		}

		updating = append(updating, obj)
	}

	/* Create a context and begin a transaction for bulk upsert */
	ctx := context.Background()
	tx, err := game.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, obj := range updating {
		_, err = tx.ExecContext(ctx, `
			UPDATE
				object_instances
			SET
				name = ?,
				short_description = ?,
				long_description = ?,
				description = ?,
				value_1 = ?,
				value_2 = ?,
				value_3 = ?,
				value_4 = ?
		`, obj.name, obj.shortDescription, obj.longDescription, obj.description, obj.value0, obj.value1, obj.value2, obj.value3)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (game *Game) LoadPlayerInventory(ch *Character) error {
	rows, err := game.db.Query(`
		SELECT
			object_instances.id,
			object_instances.parent_id,
			object_instances.name,
			object_instances.short_description,
			object_instances.long_description,
			object_instances.description,
			object_instances.value_1,
			object_instances.value_2,
			object_instances.value_3,
			object_instances.value_4
		FROM
			object_instances
		INNER JOIN
			player_character_object
		ON
			object_instances.id = player_character_object.object_instance_id
		WHERE
			player_character_object.player_character_id = ?
	`, ch.id)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		obj := &ObjectInstance{
			game:      game,
			contents:  NewLinkedList(),
			inside:    nil,
			carriedBy: nil,
		}

		err = rows.Scan(&obj.id, &obj.parentId, &obj.name, &obj.shortDescription, &obj.longDescription, &obj.description, &obj.value0, &obj.value1, &obj.value2, &obj.value3)
		if err != nil {
			return err
		}

		ch.addObject(obj)
	}

	for iter := ch.inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		rows, err := game.db.Query(`
			SELECT
				object_instances.id,
				object_instances.parent_id,
				object_instances.name,
				object_instances.short_description,
				object_instances.long_description,
				object_instances.description,
				object_instances.value_1,
				object_instances.value_2,
				object_instances.value_3,
				object_instances.value_4
			WHERE
				object_instances.inside_object_instance_id = ?
		`, obj.id)
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			containedObj := &ObjectInstance{
				game:      game,
				contents:  NewLinkedList(),
				inside:    nil,
				carriedBy: nil,
			}

			err = rows.Scan(&containedObj.id, &containedObj.parentId, &containedObj.name, &containedObj.shortDescription, &containedObj.longDescription, &containedObj.description, &containedObj.value0, &containedObj.value1, &containedObj.value2, &containedObj.value3)
			if err != nil {
				return err
			}

			obj.addObject(containedObj)
		}
	}

	return nil
}

/*
 * FindPlayerByName returns a reference to the named PC, if such an account
 * exists.  Character returned may or may not have a nullable client property.
 *
 * If the character was not already online in an active session, then attempt
 * a lookup against the database.
 */
func (game *Game) FindPlayerByName(username string) (*Character, *Room, error) {
	for client := range game.clients {
		if client.character != nil && client.character.name == username {
			return client.character, client.character.Room, nil
		}
	}

	/* There was no online player with this name, search the database. */
	row := game.db.QueryRow(`
		SELECT
			id,
			username,
			wizard,
			room_id,
			race_id,
			job_id,
			level,
			experience,
			practices,
			health,
			max_health,
			mana,
			max_mana,
			stamina,
			max_stamina,
			stat_str,
			stat_dex,
			stat_int,
			stat_wis,
			stat_con,
			stat_cha,
			stat_lck
		FROM
			player_characters
		WHERE
			username = ?
		AND
			deleted_at IS NULL
	`, username)

	ch := NewCharacter()
	ch.game = game

	var roomId uint
	var raceId uint
	var jobId uint

	err := row.Scan(&ch.id, &ch.name, &ch.wizard, &roomId, &raceId, &jobId, &ch.level, &ch.experience, &ch.practices, &ch.health, &ch.maxHealth, &ch.mana, &ch.maxMana, &ch.stamina, &ch.maxStamina, &ch.Strength, &ch.Dexterity, &ch.Intelligence, &ch.Wisdom, &ch.Constitution, &ch.Charisma, &ch.Luck)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}

		return nil, nil, err
	}

	ch.race = FindRaceByID(raceId)
	if ch.race == nil {
		return nil, nil, fmt.Errorf("failed to load race %d", raceId)
	}

	ch.job = FindJobByID(jobId)
	if ch.job == nil {
		return nil, nil, fmt.Errorf("failed to load job %d", jobId)
	}

	room, err := game.LoadRoomIndex(roomId)
	if err != nil {
		return nil, nil, err
	}

	err = game.LoadPlayerInventory(ch)
	if err != nil {
		return nil, nil, err
	}

	err = ch.LoadPlayerSkills()
	if err != nil {
		return nil, nil, err
	}

	return ch, room, nil
}

func (ch *Character) clearOutputBuffer() {
	ch.pages = make([][]byte, 1)
	ch.pages[0] = make([]byte, ch.pageSize)
	ch.pageCursor = 0
}

func (ch *Character) flushOutput() {
	defer func() {
		recover()
	}()

	for _, page := range ch.pages {
		ch.client.send <- page
	}

	ch.clearOutputBuffer()
}

func (ch *Character) gainExperience(experience int) {
	if ch.flags&CHAR_IS_PLAYER == 0 {
		return
	}

	if ch.level < LevelHero {
		ch.Send(fmt.Sprintf("{WYou gained %d experience points.{x\r\n", experience))
		ch.experience = ch.experience + uint(experience)

		/* If we gain enough experience to level up multiple times */
		for {
			if ch.level >= LevelHero {
				break
			}

			tnl := uint(ch.experienceRequiredForLevel(int(ch.level + 1)))

			if ch.experience > tnl {
				ch.level = ch.level + 1

				/* Calculate stat gains, skill points, etc... */
				healthGain := 20
				manaGain := 20
				staminaGain := 20

				ch.maxHealth += healthGain
				ch.health += healthGain
				ch.maxMana += manaGain
				ch.mana += manaGain
				ch.maxStamina += staminaGain
				ch.stamina += staminaGain

				ch.Send(fmt.Sprintf("{YYou have advanced to level %d!\r\n{x", ch.level))
				/* Any extra announce/log */

				continue
			}

			break
		}
	}
}

func (ch *Character) isFighting() bool {
	return ch.Fighting != nil
}

func (ch *Character) Write(data []byte) (n int, err error) {
	if ch.client == nil {
		/* If there is no client, succeed silently. */
		return len(data), nil
	}

	if len(data)+ch.pageCursor > ch.pageSize {
		return 0, errors.New("overflowed client buffer")
	}

	/*
	 * This will need to be rewritten; we need to divide the data length by the page size, then
	 * drain the data by chunks.  This is currently only "coincidentally working."
	 */
	copy(ch.pages[ch.pageCursor/ch.pageSize][ch.pageCursor:ch.pageCursor+len(data)], data[:])
	ch.pageCursor = ch.pageCursor + len(data)

	return len(data), nil
}

/*
 * TODO: implement validation logic restricting silly/invalid/breaking names.
 */
func (game *Game) IsValidPCName(name string) bool {
	/* Length bounds */
	if len(name) < 3 || len(name) > 14 || name == UnauthenticatedUsername {
		return false
	}

	/* If any character is non-alpha, invalidate. */
	for c := range name {
		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') {
			return false
		}
	}

	return true
}

func (ch *Character) Send(text string) {
	var output string = string(text)

	if ch.client != nil {
		output = ch.client.TranslateColourCodes(output)
	}

	_, err := ch.Write([]byte(output))
	if err != nil {
		log.Printf("Failed to write to character: %v.\r\n", err)
		return
	}
}

func (ch *Character) getMaxItemsInventory() int {
	return 20
}

func (ch *Character) getMaxCarryWeight() float64 {
	return 200.0
}

func (ch *Character) getShortDescription(viewer *Character) string {
	if ch.flags&CHAR_IS_PLAYER != 0 {
		return ch.name
	}

	return ch.shortDescription
}

func (ch *Character) getShortDescriptionUpper(viewer *Character) string {
	var short string = ch.getShortDescription(viewer)

	if short == "" {
		return ""
	}

	runes := []rune(short)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (ch *Character) getLongDescription(viewer *Character) string {
	if ch.Fighting != nil {
		if viewer == ch.Fighting {
			return fmt.Sprintf("%s is here, fighting you!", ch.getShortDescriptionUpper(viewer))
		}

		return fmt.Sprintf("%s is here, fighting %s.", ch.getShortDescriptionUpper(viewer), ch.Fighting.getShortDescription(viewer))
	}

	if ch.flags&CHAR_IS_PLAYER != 0 {
		return fmt.Sprintf("%s is here.", ch.name)
	}

	return ch.longDescription
}

func (ch *Character) addObject(obj *ObjectInstance) {
	ch.inventory.Insert(obj)

	obj.carriedBy = ch
	obj.inRoom = nil
	obj.inside = nil
}

func (ch *Character) removeObject(obj *ObjectInstance) {
	ch.inventory.Remove(obj)

	obj.carriedBy = nil
}

func (ch *Character) findObjectInRoom(argument string) *ObjectInstance {
	processed := strings.ToLower(argument)

	if ch.Room == nil || len(processed) < 1 {
		return nil
	}

	for iter := ch.Room.objects.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		nameParts := strings.Split(obj.name, " ")
		for _, part := range nameParts {
			if strings.Compare(strings.ToLower(part), processed) == 0 {
				return obj
			}
		}
	}

	return nil
}

func (ch *Character) findObjectOnSelf(argument string) *ObjectInstance {
	processed := strings.ToLower(argument)

	if ch.Room == nil || len(processed) < 1 {
		return nil
	}

	for iter := ch.inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		nameParts := strings.Split(obj.name, " ")
		for _, part := range nameParts {
			if strings.Compare(strings.ToLower(part), processed) == 0 {
				return obj
			}
		}
	}

	return nil
}

func (ch *Character) findCharacterInRoom(argument string) *Character {
	processed := strings.ToLower(argument)

	if processed == "self" {
		return ch
	}

	if ch.Room == nil || len(processed) < 1 {
		return nil
	}

	for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		nameParts := strings.Split(rch.name, " ")
		for _, part := range nameParts {
			if strings.Compare(strings.ToLower(part), processed) == 0 {
				return rch
			}
		}
	}

	return nil
}

func (game *Game) Broadcast(message string) {
	log.Printf("Broadcast: %s\r\n", message)
	for iter := game.characters.Head; iter != nil; iter = iter.Next {
		ch := iter.Value.(*Character)

		ch.Send(message)
	}
}

func NewCharacter() *Character {
	character := &Character{}

	character.id = -1
	character.wizard = false
	character.afk = nil
	character.job = nil
	character.flags = 0
	character.Fighting = nil
	character.Combat = nil
	character.race = nil
	character.Room = nil
	character.practices = 0
	character.pageSize = 4096
	character.position = PositionDead
	character.pages = make([][]byte, 1)
	character.pages[0] = make([]byte, character.pageSize)
	character.pageCursor = 0

	character.name = UnauthenticatedUsername
	character.client = nil
	character.level = 0
	character.experience = 0
	character.inventory = NewLinkedList()
	character.skills = make(map[uint]*Proficiency)

	character.Strength = 10
	character.Dexterity = 10
	character.Intelligence = 10
	character.Wisdom = 10
	character.Constitution = 10
	character.Charisma = 10
	character.Luck = 10

	character.equipment = make([]*ObjectInstance, WearLocationMax)
	return character
}
