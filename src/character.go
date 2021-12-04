/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/dop251/goja"
	"golang.org/x/crypto/bcrypt"
)

const UnauthenticatedUsername = "unnamed"

type Job struct {
	Id                         uint        `json:"id"`
	Name                       string      `json:"name"`
	DisplayName                string      `json:"display_name"`
	Playable                   bool        `json:"playable"`
	ExperienceRequiredModifier float64     `json:"experience_required_modifier"`
	Skills                     *LinkedList `json:"skills"`
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
	CHAR_TRAIN      = 1 << 4
	CHAR_PRACTICE   = 1 << 5
	CHAR_HEALER     = 1 << 6
	CHAR_SHOPKEEPER = 1 << 7
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
	AFFECT_SANCTUARY    = 1
	AFFECT_HASTE        = 1 << 1
	AFFECT_SLOW         = 1 << 2
	AFFECT_POISON       = 1 << 3
	AFFECT_SILENCE      = 1 << 4
	AFFECT_DETECT_MAGIC = 1 << 5
	AFFECT_FIRESHIELD   = 1 << 6
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
	Game   *Game   `json:"game"`
	Client *Client `json:"client"`

	Inventory *LinkedList `json:"inventory"`

	output       []byte
	outputCursor int
	outputHead   int
	outputLines  int
	inputCursor  int

	PlaneIndex *Point          `json:"planeIndex"`
	Room       *Room           `json:"room"`
	Combat     *Combat         `json:"combat"`
	Fighting   *Character      `json:"fighting"`
	Casting    *CastingContext `json:"casting"`

	Following *Character  `json:"following"`
	Leader    *Character  `json:"leader"`
	Group     *LinkedList `json:"group"`

	Id int `json:"id"`

	Name             string `json:"name"`
	ShortDescription string `json:"shortDescription"`
	LongDescription  string `json:"longDescription"`
	Description      string `json:"description"`

	Wizard bool  `json:"wizard"`
	Wiznet bool  `json:"wiznet"`
	Job    *Job  `json:"job"`
	Race   *Race `json:"race"`

	Level      uint `json:"level"`
	Experience uint `json:"experience"`
	Practices  int  `json:"practices"`

	Affected int                   `json:"affected"`
	Effects  *LinkedList           `json:"effects"`
	Skills   map[uint]*Proficiency `json:"skills"`

	Gold  int               `json:"gold"`
	Flags int               `json:"flags"`
	Afk   *AwayFromKeyboard `json:"afk"`

	Health     int `json:"health"`
	MaxHealth  int `json:"maxHealth"`
	Mana       int `json:"mana"`
	MaxMana    int `json:"maxMana"`
	Stamina    int `json:"stamina"`
	MaxStamina int `json:"maxStamina"`

	Position int

	Strength     int `json:"strength"`
	Dexterity    int `json:"dexterity"`
	Intelligence int `json:"intelligence"`
	Wisdom       int `json:"wisdom"`
	Constitution int `json:"constitution"`
	Charisma     int `json:"charisma"`
	Luck         int `json:"luck"`

	Defense int

	temporaryHash string
}

func (ch *Character) IsEqual(och *Character) bool {
	return och == ch
}

func (ch *Character) experienceRequiredForLevel(level int) int {
	required := int(500*(level*level) - (500 * level))

	if ch.Job != nil {
		required = int(float64(required) * ch.Job.ExperienceRequiredModifier)
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

func (ch *Character) onUpdate() {
	/*
	 * Regenerate some health and mana every tick if not in a room with ROOM_EVIL_AURA set.
	 * Always regenerate some stamina.
	 */
	if ch.Room == nil || ch.Room.Flags&ROOM_EVIL_AURA == 0 {
		if ch.Health < ch.MaxHealth {
			ch.Health = int(math.Min(float64(ch.MaxHealth), float64(ch.Health+3)))
		}

		if ch.Mana < ch.MaxMana {
			ch.Mana = int(math.Min(float64(ch.MaxMana), float64(ch.Mana+7)))
		}
	}

	if ch.Stamina < ch.MaxStamina {
		ch.Stamina = int(math.Min(float64(ch.MaxStamina), float64(ch.Stamina+40)))
	}
}

func (ch *Character) Finalize() error {
	if ch.Client == nil || ch.Game == nil {
		/* If somehow an NPC were to try to save, do not allow it. */
		return nil
	}

	result, err := ch.Game.db.Exec(`
		INSERT INTO
			player_characters(username, password_hash, wizard, room_id, race_id, job_id, level, gold, experience, practices, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ch.Name, ch.temporaryHash, 0, RoomLimbo, ch.Race.Id, ch.Job.Id, ch.Level, ch.Gold, ch.Experience, ch.Practices, ch.Health, ch.MaxHealth, ch.Mana, ch.MaxMana, ch.Stamina, ch.MaxStamina, ch.Strength, ch.Dexterity, ch.Intelligence, ch.Wisdom, ch.Constitution, ch.Charisma, ch.Luck)
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

	ch.Id = int(userId)

	limbo, err := ch.Game.LoadRoomIndex(RoomLimbo)
	if err != nil {
		return err
	}

	ch.Room = limbo
	return nil
}

func (ch *Character) SavePlayerSkills() error {
	var proficiencyValues strings.Builder

	if len(ch.Skills) == 0 {
		return nil
	}

	for _, proficiency := range ch.Skills {
		proficiencyValues.WriteString(fmt.Sprintf("(%d, %d, %d, %d, %d),", proficiency.Id, ch.Id, proficiency.SkillId, proficiency.Job.Id, proficiency.Proficiency))
	}

	proficiencyValuesString := strings.TrimRight(proficiencyValues.String(), ",")
	_, err := ch.Game.db.Exec(fmt.Sprintf(`
	INSERT INTO
		pc_skill_proficiency (id, player_character_id, skill_id, job_id, proficiency)
	VALUES
		%s
	ON DUPLICATE KEY
		UPDATE
			proficiency = VALUES(proficiency)`,
		proficiencyValuesString))
	if err != nil {
		return err
	}

	return nil
}

func (ch *Character) Save() bool {
	if ch.Client == nil || ch.Game == nil {
		/* If somehow an NPC were to try to save, do not allow it. */
		return false
	}

	var roomId uint = RoomLimbo
	if ch.Room != nil {
		roomId = ch.Room.Id
	}
	result, err := ch.Game.db.Exec(`
		UPDATE
			player_characters
		SET
			wizard = ?,
			room_id = ?,
			race_id = ?,
			job_id = ?,
			level = ?,
			gold = ?,
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
	`, ch.Wizard, roomId, ch.Race.Id, ch.Job.Id, ch.Level, ch.Gold, ch.Experience, ch.Practices, ch.Health, ch.MaxHealth, ch.Mana, ch.MaxMana, ch.Stamina, ch.MaxStamina, ch.Strength, ch.Dexterity, ch.Intelligence, ch.Wisdom, ch.Constitution, ch.Charisma, ch.Luck, ch.Id)
	if err != nil {
		log.Printf("Failed to save character: %v.\r\n", err)
		return false
	}

	_, err = result.RowsAffected()
	if err != nil {
		log.Printf("Failed to retrieve number of rows affected: %v.\r\n", err)
		return false
	}

	err = ch.SavePlayerSkills()
	if err != nil {
		log.Printf("Failed to save player proficiencies: %v.\r\n", err)
		return false
	}

	err = ch.Game.SavePlayerInventory(ch)
	if err != nil {
		log.Printf("Failed to save player inventory: %v.\r\n", err)
		return false
	}

	return true
}

func (ch *Character) attachObject(obj *ObjectInstance) error {
	err := obj.reify()
	if err != nil {
		return err
	}

	_, err = ch.Game.db.Exec(`
	INSERT INTO
		player_character_object(player_character_id, object_instance_id)
	VALUES
		(?, ?)
	`, ch.Id, obj.Id)
	if err != nil {
		return err
	}

	return nil
}

func (ch *Character) DetachObject(obj *ObjectInstance) error {
	result, err := ch.Game.db.Exec(`
		DELETE FROM
			player_character_object
		WHERE
			player_character_id = ?
		AND
			object_instance_id = ?`, ch.Id, obj.Id)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	return nil
}

func (game *Game) SavePlayerInventory(ch *Character) error {
	/* Object instances whose records have dirtied */
	var updating []*ObjectInstance = make([]*ObjectInstance, 0)

	/* Iterate over all objects in this player's inventory */
	for iter := ch.Inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		/* If this is a container, ensure that all contained object instances are also updated */
		if obj.Contents != nil && obj.Contents.Count > 0 {
			for containerIter := obj.Contents.Head; containerIter != nil; containerIter = containerIter.Next {
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
		if obj.Id == 0 {
			continue
		}

		_, err = tx.ExecContext(ctx, `
			UPDATE
				object_instances
			SET
				name = ?,
				short_description = ?,
				long_description = ?,
				description = ?,
				wear_location = ?,
				flags = ?,
				value_1 = ?,
				value_2 = ?,
				value_3 = ?,
				value_4 = ?
			WHERE
				id = ?
		`, obj.Name, obj.ShortDescription, obj.LongDescription, obj.Description, obj.WearLocation, obj.Flags, obj.Value0, obj.Value1, obj.Value2, obj.Value3, obj.Id)
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
			object_instances.flags,
			object_instances.item_type,
			object_instances.wear_location,
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
	`, ch.Id)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		obj := &ObjectInstance{
			Game:         game,
			Contents:     NewLinkedList(),
			Inside:       nil,
			CarriedBy:    nil,
			CreatedAt:    time.Now(),
			WearLocation: -1,
		}

		err = rows.Scan(&obj.Id, &obj.ParentId, &obj.Name, &obj.ShortDescription, &obj.LongDescription, &obj.Description, &obj.Flags, &obj.ItemType, &obj.WearLocation, &obj.Value0, &obj.Value1, &obj.Value2, &obj.Value3)
		if err != nil {
			return err
		}

		ch.addObject(obj)
	}

	for iter := ch.Inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		rows, err := game.db.Query(`
			SELECT
				object_instances.id,
				object_instances.parent_id,
				object_instances.name,
				object_instances.short_description,
				object_instances.long_description,
				object_instances.description,
				object_instances.flags,
				object_instances.item_type,
				object_instances.value_1,
				object_instances.value_2,
				object_instances.value_3,
				object_instances.value_4
			FROM
				object_instances
			WHERE
				object_instances.inside_object_instance_id = ?
		`, obj.Id)
		if err != nil {
			return err
		}

		defer rows.Close()

		for rows.Next() {
			containedObj := &ObjectInstance{
				Game:         game,
				Contents:     NewLinkedList(),
				Inside:       nil,
				CarriedBy:    nil,
				CreatedAt:    time.Now(),
				WearLocation: -1,
			}

			err = rows.Scan(&containedObj.Id, &containedObj.ParentId, &containedObj.Name, &containedObj.ShortDescription, &containedObj.LongDescription, &containedObj.Description, &containedObj.Flags, &containedObj.ItemType, &containedObj.Value0, &containedObj.Value1, &containedObj.Value2, &containedObj.Value3)
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
		if client.ConnectionState == ConnectionStatePlaying && client.Character != nil && client.Character.Name == username {
			return client.Character, client.Character.Room, nil
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
			gold,
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
	ch.Game = game

	var roomId uint
	var raceId uint
	var jobId uint

	err := row.Scan(&ch.Id, &ch.Name, &ch.Wizard, &roomId, &raceId, &jobId, &ch.Level, &ch.Gold, &ch.Experience, &ch.Practices, &ch.Health, &ch.MaxHealth, &ch.Mana, &ch.MaxMana, &ch.Stamina, &ch.MaxStamina, &ch.Strength, &ch.Dexterity, &ch.Intelligence, &ch.Wisdom, &ch.Constitution, &ch.Charisma, &ch.Luck)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, nil
		}

		return nil, nil, err
	}

	ch.Race = FindRaceByID(raceId)
	if ch.Race == nil {
		return nil, nil, fmt.Errorf("failed to load race %d", raceId)
	}

	ch.Job = FindJobByID(jobId)
	if ch.Job == nil {
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
	ch.output = make([]byte, 32768)
	ch.outputHead = 0
	ch.outputCursor = 0
	ch.inputCursor = DefaultMaxLines
}

func (ch *Character) flushOutput() {
	if ch.output[0] == 0 {
		return
	}

	var page bytes.Buffer
	var lines []string
	var maxLines int = DefaultMaxLines

	scan := bufio.NewScanner(strings.NewReader(string(ch.output)))
	scan.Split(func(data []byte, eof bool) (advance int, token []byte, err error) {
		if eof && len(data) == 0 {
			return 0, nil, nil
		}

		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			if len(data) > 0 && data[len(data)-1] == '\r' {
				return i + 1, data[0 : len(data)-1], nil
			}

			return i + 1, data[0:i], nil
		}

		if eof {
			if len(data) > 0 && data[len(data)-1] == '\r' {
				return len(data), data[:len(data)-1], nil
			}

			return len(data), data, nil
		}

		return 0, nil, nil
	})

	for scan.Scan() {
		lines = append(lines, scan.Text())
	}

	ch.outputLines = len(lines)
	if ch.outputLines <= 1 {
		return
	}

	if len(lines) <= maxLines {
		ch.Client.send <- ch.output
		ch.clearOutputBuffer()
		return
	}

	for index := ch.outputCursor; index <= ch.inputCursor; index++ {
		if index >= len(lines) {
			ch.Client.send <- page.Bytes()
			ch.clearOutputBuffer()
			return
		}

		if index-ch.outputCursor >= maxLines {
			ch.Client.send <- page.Bytes()
			ch.outputCursor += maxLines

			amount := ch.outputCursor * 100 / ch.outputLines

			ch.Client.send <- []byte(fmt.Sprintf("[ Press return to continue (%d%%) ]\r\n", amount))
			return
		}

		page.Write([]byte(lines[index]))
		page.Write([]byte("\r\n"))
	}
}

func (ch *Character) gainExperience(experience int) {
	if ch.Flags&CHAR_IS_PLAYER == 0 {
		return
	}

	if ch.Level < LevelHero {
		ch.Send(fmt.Sprintf("{WYou gained %d experience points.{x\r\n", experience))
		ch.Experience = ch.Experience + uint(experience)

		/* If we gain enough experience to level up multiple times */
		for {
			if ch.Level >= LevelHero {
				break
			}

			tnl := uint(ch.experienceRequiredForLevel(int(ch.Level + 1)))

			if ch.Experience > tnl {
				ch.Level = ch.Level + 1

				/* Calculate stat gains, skill points, etc... */
				healthGain := 20
				manaGain := 20
				staminaGain := 20
				practicesGain := rand.Intn(10)

				ch.MaxHealth += healthGain
				ch.Health += healthGain
				ch.MaxMana += manaGain
				ch.Mana += manaGain
				ch.MaxStamina += staminaGain
				ch.Stamina += staminaGain
				ch.Practices += practicesGain

				ch.Send(fmt.Sprintf("{YYou have advanced to level %d!\r\n{x", ch.Level))
				ch.Send(fmt.Sprintf("{WOh yeah! You gained %d hp, %d mana, %d stamina, and %d practice sessions.{x\r\n", healthGain, manaGain, staminaGain, practicesGain))

				err := ch.syncJobSkills()
				if err != nil {
					log.Println(err)
				}

				ch.Save()
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
	if ch.Client == nil {
		/* If there is no client, succeed silently. */
		return len(data), nil
	}

	if ch.outputHead+len(data) >= 32768 {
		/* Clear buffer and drop connection for overflowing the pager */
		ch.clearOutputBuffer()
		ch.Client.close <- true
		ch.Client.conn.Close()
		return 0, nil
	}

	copy(ch.output[ch.outputHead:ch.outputHead+len(data)], data[:])
	ch.outputHead = ch.outputHead + len(data)

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
	for _, c := range name {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') {
			return false
		}
	}

	return true
}

func (ch *Character) Send(text string) {
	var output string = string(text)

	if ch.Client != nil {
		output = ch.Client.TranslateColourCodes(output)
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

func (ch *Character) GetShortDescription(viewer *Character) string {
	if ch.Flags&CHAR_IS_PLAYER != 0 {
		return ch.Name
	}

	return ch.ShortDescription
}

func (ch *Character) GetShortDescriptionUpper(viewer *Character) string {
	var short string = ch.GetShortDescription(viewer)

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
			return fmt.Sprintf("%s is here, fighting you!", ch.GetShortDescriptionUpper(viewer))
		}

		return fmt.Sprintf("%s is here, fighting %s.", ch.GetShortDescriptionUpper(viewer), ch.Fighting.GetShortDescription(viewer))
	}

	if ch.Flags&CHAR_IS_PLAYER != 0 {
		return fmt.Sprintf("%s is here.", ch.Name)
	}

	return ch.LongDescription
}

func (ch *Character) addObject(obj *ObjectInstance) {
	ch.Inventory.Insert(obj)

	obj.CarriedBy = ch
	obj.InRoom = nil
	obj.Inside = nil
}

func (ch *Character) RemoveObject(obj *ObjectInstance) {
	ch.Inventory.Remove(obj)

	obj.CarriedBy = nil
}

func (obj *ObjectInstance) findObjectInSelf(ch *Character, argument string) *ObjectInstance {
	processed := strings.ToLower(argument)

	if ch.Room == nil || len(processed) < 1 || obj.Contents == nil || obj.Contents.Count < 1 {
		return nil
	}

	for iter := obj.Contents.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		nameParts := strings.Split(obj.Name, " ")
		for _, part := range nameParts {
			if strings.Compare(strings.ToLower(part), processed) == 0 {
				return obj
			}
		}
	}

	return nil
}

func (ch *Character) findObjectInRoom(argument string) *ObjectInstance {
	processed := strings.ToLower(argument)

	if ch.Room == nil || len(processed) < 1 {
		return nil
	}

	for iter := ch.Room.Objects.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		nameParts := strings.Split(obj.Name, " ")
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

	for iter := ch.Inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		if obj.WearLocation != -1 {
			continue
		}

		nameParts := strings.Split(obj.Name, " ")
		for _, part := range nameParts {
			if strings.Compare(strings.ToLower(part), processed) == 0 {
				return obj
			}
		}
	}

	return nil
}

func (ch *Character) FindCharacterInRoom(argument string) *Character {
	processed := strings.ToLower(argument)

	if processed == "self" {
		return ch
	}

	if ch.Room == nil || len(processed) < 1 {
		return nil
	}

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		nameParts := strings.Split(rch.Name, " ")
		for _, part := range nameParts {
			if strings.Compare(strings.ToLower(part), processed) == 0 {
				return rch
			}
		}
	}

	return nil
}

func (game *Game) Broadcast(message string, filter goja.Callable) {
	var recipients []*Character = make([]*Character, 0)

	for iter := game.Characters.Head; iter != nil; iter = iter.Next {
		var result bool = false

		ch := iter.Value.(*Character)

		if filter != nil {
			val, err := filter(game.vm.ToValue(ch))
			if err != nil {
				log.Printf("Room.Broadcast failed: %v\r\n", err)
				break
			}

			result = val.ToBoolean()
		}

		if result || filter == nil {
			recipients = append(recipients, ch)
		}
	}

	/* Send message to gathered users */
	for _, rcpt := range recipients {
		rcpt.Send(message)
	}
}

func WiznetBroadcastFilter(ch *Character) bool {
	return ch.Wiznet
}

func (game *Game) broadcast(message string, filterFn func(*Character) bool) {
	var recipients []*Character = make([]*Character, 0)

	for iter := game.Characters.Head; iter != nil; iter = iter.Next {
		var result bool = false

		ch := iter.Value.(*Character)

		if filterFn != nil {
			result = filterFn(ch)
		}

		if result || filterFn == nil {
			recipients = append(recipients, ch)
		}
	}

	/* Send message to gathered users */
	for _, rcpt := range recipients {
		rcpt.Send(message)
	}
}

func NewCharacter() *Character {
	character := &Character{}

	character.Id = -1
	character.Wizard = false
	character.Afk = nil
	character.Job = nil
	character.Flags = 0
	character.Gold = 0
	character.Fighting = nil
	character.Combat = nil
	character.Race = nil
	character.Room = nil
	character.PlaneIndex = nil
	character.Wiznet = false
	character.Practices = 0
	character.Position = PositionDead
	character.output = make([]byte, 65536)
	character.outputCursor = 0
	character.inputCursor = DefaultMaxLines
	character.outputHead = 0

	character.Affected = 0
	character.Name = UnauthenticatedUsername
	character.Client = nil
	character.Level = 0
	character.Experience = 0
	character.Inventory = NewLinkedList()
	character.Effects = NewLinkedList()
	character.Skills = make(map[uint]*Proficiency)

	character.Defense = 0
	character.Strength = 10
	character.Dexterity = 10
	character.Intelligence = 10
	character.Wisdom = 10
	character.Constitution = 10
	character.Charisma = 10
	character.Luck = 10

	return character
}
