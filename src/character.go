/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const UnauthenticatedUsername = "unnamed"

type Job struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Playable    bool   `json:"playable"`
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

/*
 * This character structure is shared by both player-characters (human beings
 * connected through a session instance available via the client pointer.)
 */
type Character struct {
	client    *Client
	inventory *LinkedList

	pages      [][]byte
	pageSize   int
	pageCursor int

	room     *Room
	combat   *Combat
	fighting *Character

	id int

	name             string
	shortDescription string
	longDescription  string
	description      string

	wizard     bool
	job        *Job
	race       *Race
	level      uint
	experience uint

	flags int
	afk   *AwayFromKeyboard

	health     int
	maxHealth  int
	mana       int
	maxMana    int
	stamina    int
	maxStamina int

	strength     int
	dexterity    int
	intelligence int
	wisdom       int
	constitution int
	charisma     int
	luck         int

	temporaryHash string
}

func ExperienceRequiredForLevel(level int) int {
	return int(500*(level*level) - (500 * level))
}

/*
 * FindPlayerByName returns a reference to the named PC, if such an account
 * exists.  Character returned may or may not have a nullable client property.
 *
 * If the character was not already online in an active session, then attempt
 * a lookup against the database.
 */
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
	/* Regen, script hooks, etc. */
	if ch.health < ch.maxHealth {
		ch.health = int(math.Min(float64(ch.maxHealth), float64(ch.health+5)))
	}
}

func (ch *Character) Finalize() bool {
	if ch.client == nil || ch.client.game == nil {
		/* If somehow an NPC were to try to save, do not allow it. */
		return false
	}

	result, err := ch.client.game.db.Exec(`
		INSERT INTO
			player_characters(username, password_hash, wizard, room_id, race_id, job_id, level, experience, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ch.name, ch.temporaryHash, 0, RoomLimbo, ch.race.Id, ch.job.Id, ch.level, ch.experience, ch.health, ch.maxHealth, ch.mana, ch.maxMana, ch.stamina, ch.maxStamina, ch.strength, ch.dexterity, ch.intelligence, ch.wisdom, ch.constitution, ch.charisma, ch.luck)
	ch.temporaryHash = ""
	if err != nil {
		log.Printf("Failed to finalize new character: %v.\r\n", err)
		return false
	}

	userId, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve insert id: %v.\r\n", err)
		return false
	}

	ch.id = int(userId)

	limbo, err := ch.client.game.LoadRoomIndex(RoomLimbo)
	if err != nil {
		return false
	}

	ch.room = limbo
	return true
}

func (ch *Character) Save() bool {
	if ch.client == nil || ch.client.game == nil {
		/* If somehow an NPC were to try to save, do not allow it. */
		return false
	}

	var roomId uint = RoomLimbo
	if ch.room != nil {
		roomId = ch.room.id
	}
	result, err := ch.client.game.db.Exec(`
		UPDATE
			player_characters
		SET
			wizard = ?,
			room_id = ?,
			race_id = ?,
			job_id = ?,
			level = ?,
			experience = ?,
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
	`, ch.wizard, roomId, ch.race.Id, ch.job.Id, ch.level, ch.experience, ch.health, ch.maxHealth, ch.mana, ch.maxMana, ch.stamina, ch.maxStamina, ch.strength, ch.dexterity, ch.intelligence, ch.wisdom, ch.constitution, ch.charisma, ch.luck, ch.id)
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
			return client.character, client.character.room, nil
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

	var roomId uint
	var raceId uint
	var jobId uint

	err := row.Scan(&ch.id, &ch.name, &ch.wizard, &roomId, &raceId, &jobId, &ch.level, &ch.experience, &ch.health, &ch.maxHealth, &ch.mana, &ch.maxMana, &ch.stamina, &ch.maxStamina, &ch.strength, &ch.dexterity, &ch.intelligence, &ch.wisdom, &ch.constitution, &ch.charisma, &ch.luck)

	/* Sanity check for pointers by race and job id before continuing */
	_, ok := RaceTable[raceId]
	if !ok {
		return nil, nil, nil
	}

	_, ok = JobTable[jobId]
	if !ok {
		return nil, nil, nil
	}

	ch.race = RaceTable[raceId]
	ch.job = JobTable[jobId]

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}

		return nil, nil, err
	}

	room, err := game.LoadRoomIndex(roomId)
	if err != nil {
		return nil, nil, err
	}

	return ch, room, nil
}

func (ch *Character) flushOutput() {
	for _, page := range ch.pages {
		ch.client.send <- page
	}

	ch.pages = make([][]byte, 1)
	ch.pages[0] = make([]byte, ch.pageSize)
	ch.pageCursor = 0
}

func (ch *Character) gainExperience(experience int) {
	if ch.flags&CHAR_IS_PLAYER == 0 {
		return
	}

	ch.Send(fmt.Sprintf("{WYou gained %d experience points.{x\r\n", experience))
	ch.experience = ch.experience + uint(experience)

	if ch.level < LevelHero {
		/* If we gain enough experience to level up multiple times */
		for {
			if ch.level >= LevelHero {
				break
			}

			tnl := uint(ExperienceRequiredForLevel(int(ch.level + 1)))

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
	return ch.fighting != nil
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
	if ch.fighting != nil {
		if viewer == ch.fighting {
			return fmt.Sprintf("%s is here, fighting you!", ch.getShortDescriptionUpper(viewer))
		}

		return fmt.Sprintf("%s is here, fighting %s.", ch.getShortDescriptionUpper(viewer), ch.fighting.getShortDescription(viewer))
	}

	if ch.flags&CHAR_IS_PLAYER != 0 {
		return fmt.Sprintf("%s is here.", ch.name)
	}

	return ch.longDescription
}

func (ch *Character) getLongDescriptionUpper(viewer *Character) string {
	var long string = ch.getLongDescription(viewer)

	if long == "" {
		return ""
	}

	runes := []rune(long)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (ch *Character) addObject(obj *ObjectInstance) {
	ch.inventory.Insert(obj)
}

func (ch *Character) removeObject(obj *ObjectInstance) {
	ch.inventory.Remove(obj)
}

func NewCharacter() *Character {
	character := &Character{}

	character.id = -1
	character.wizard = false
	character.afk = nil
	character.job = nil
	character.flags = 0
	character.fighting = nil
	character.combat = nil
	character.race = nil
	character.room = nil
	character.pageSize = 2048
	character.pages = make([][]byte, 1)
	character.pages[0] = make([]byte, character.pageSize)
	character.pageCursor = 0

	character.name = UnauthenticatedUsername
	character.client = nil
	character.level = 0
	character.experience = 0
	character.inventory = NewLinkedList()

	character.strength = 10
	character.dexterity = 10
	character.intelligence = 10
	character.wisdom = 10
	character.constitution = 10
	character.charisma = 10
	character.luck = 10

	return character
}
