/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"time"
)

func (ch *Character) getHealthFeedback(viewer *Character) string {
	healthPercentage := ch.Health * 100 / ch.MaxHealth

	if healthPercentage >= 100 {
		return fmt.Sprintf("{G%s{G is in perfect health.{x", ch.GetShortDescriptionUpper(viewer))
	} else if healthPercentage >= 80 {
		return fmt.Sprintf("{g%s{g is barely scratched.{x", ch.GetShortDescriptionUpper(viewer))
	} else if healthPercentage >= 60 {
		return fmt.Sprintf("{w%s{w has several cuts and scratches.{x", ch.GetShortDescriptionUpper(viewer))
	} else if healthPercentage >= 40 {
		return fmt.Sprintf("{Y%s{Y has quite a few wounds.{x", ch.GetShortDescriptionUpper(viewer))
	} else if healthPercentage >= 25 {
		return fmt.Sprintf("{M%s{M looks pretty hurt.{x", ch.GetShortDescriptionUpper(viewer))
	} else if healthPercentage >= 10 {
		return fmt.Sprintf("{R%s{R is in awful condition.{x", ch.GetShortDescriptionUpper(viewer))
	} else {
		return fmt.Sprintf("{D%s{D is about to die.{x", ch.GetShortDescriptionUpper(viewer))
	}
}

func (ch *Character) examineCharacter(other *Character) {
	if other.Flags&CHAR_IS_PLAYER == 0 {
		ch.Send(fmt.Sprintf("{G%s{x\r\n", other.Description))
	}

	ch.Send(fmt.Sprintf("%s\r\n", other.getHealthFeedback(ch)))

	for i := WearLocationNone + 1; i < WearLocationMax; i++ {
		var obj *ObjectInstance = other.getEquipment(i)

		if obj == nil {
			continue
		}

		ch.Send(fmt.Sprintf("{C%s{x%s{x\r\n", WearLocations[i], obj.GetShortDescription(ch)))
	}

	peek := ch.FindProficiencyByName("peek")
	if peek != nil && rand.Intn(100) < peek.Proficiency {
		if other.Inventory.Count > 0 {
			ch.Send(fmt.Sprintf("{Y%s{Y is carrying the following items:{x\r\n", other.GetShortDescriptionUpper(ch)))
			ch.listObjects(other.Inventory, false, true)
			return
		}
	}
}

/* List all commands available to the player in rows of 7 items. */
func do_help(ch *Character, arguments string) {
	var buf strings.Builder
	var index int = 0

	var commands []string = []string{}

	for _, command := range CommandTable {
		found := false

		for _, c := range commands {
			if c == command.Name {
				found = true
			}
		}

		if !found {
			commands = append(commands, command.Name)
		}
	}

	sort.Strings(commands)

	for _, command := range commands {
		if ch.Level <= CommandTable[command].MinimumLevel || CommandTable[command].Hidden {
			continue
		}

		buf.WriteString(fmt.Sprintf("%-10s ", command))
		index++

		if index%7 == 0 {
			buf.WriteString("\r\n")
		}
	}

	if index%7 != 0 {
		buf.WriteString("\r\n")
	}

	ch.Send(buf.String())
}

/* Display relevant game information about the player's character. */
func do_score(ch *Character, arguments string) {
	var buf strings.Builder

	healthPercentage := ch.Health * 100 / ch.MaxHealth
	manaPercentage := ch.Mana * 100 / ch.MaxMana
	staminaPercentage := ch.Stamina * 100 / ch.MaxStamina

	currentHealthColour := SeverityColourFromPercentage(healthPercentage)
	currentManaColour := SeverityColourFromPercentage(manaPercentage)
	currentStaminaColour := SeverityColourFromPercentage(staminaPercentage)

	buf.WriteString("\r\n{D┌─ {WCharacter Information {D──────────────────┬─ {WStatistics{D ───────┐{x\r\n")
	buf.WriteString(fmt.Sprintf("{D│ {CName:    {c%-13s                   {D│ Strength:       {M%2d{D │\r\n", ch.Name, ch.Strength))
	if ch.Level < LevelHero {
		buf.WriteString(fmt.Sprintf("{D│ {CLevel:   {c%-3d  {D[%8d exp. until next] {D│ Dexterity:      {M%2d{D │\r\n", ch.Level, ch.experienceRequiredForLevel(int(ch.Level+1))-int(ch.Experience), ch.Dexterity))
	} else {
		buf.WriteString(fmt.Sprintf("{D│ {CLevel:   {c%-3d                             {D│ Dexterity:      {M%2d{D │\r\n", ch.Level, ch.Dexterity))
	}
	buf.WriteString(fmt.Sprintf("{D│ {CRace:    {c%-21s           {D│ Intelligence:   {M%2d{D │\r\n", ch.Race.DisplayName, ch.Intelligence))
	buf.WriteString(fmt.Sprintf("{D│ {CJob:     {c%-21s           {D│ Wisdom:         {M%2d{D │\r\n", ch.Job.DisplayName, ch.Wisdom))
	buf.WriteString(fmt.Sprintf("{D│ {CHealth:  {c%s%-20s                {D│ Constitution:   {M%2d{D │\r\n",
		currentHealthColour,
		fmt.Sprintf("%-5d{w/{G%-5d",
			ch.Health,
			ch.MaxHealth),
		ch.Constitution))
	buf.WriteString(fmt.Sprintf("{D│ {CMana:    {c%s%-18s                  {D│ Charisma:       {M%2d{D │\r\n",
		currentManaColour,
		fmt.Sprintf("%-5d{w/{G%-5d",
			ch.Mana,
			ch.MaxMana),
		ch.Charisma))
	buf.WriteString(fmt.Sprintf("{D│ {CStamina: {c%s%-21s               {D│ Luck:           {M%2d{D │\r\n",
		currentStaminaColour,
		fmt.Sprintf("%-5d{w/{G%-5d",
			ch.Stamina,
			ch.MaxStamina),
		ch.Luck))
	buf.WriteString("{D└──────────────────────────────────────────┴────────────────────┘{x\r\n")

	output := buf.String()
	ch.Send(output)
}

/* Display a list of players online (and visible to the current player character!) */
func do_who(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("\r\n{CThe following players are online:{x\r\n")

	characters := make([]*Character, 0)

	for client := range ch.Game.clients {
		if client.character != nil && client.connectionState >= ConnectionStatePlaying {
			characters = append(characters, client.character)
		}
	}

	sort.Slice(characters, func(i int, j int) bool {
		return characters[i].Level > characters[j].Level
	})

	for _, character := range characters {
		var flagsString strings.Builder

		if character.Afk != nil {
			afkMinutes := int(time.Since(character.Afk.startedAt).Minutes())

			flagsString.WriteString(fmt.Sprintf("{G[AFK %dm]{x ", afkMinutes))
		}

		jobDisplay := character.Job.DisplayName
		if character.Level == LevelAdmin {
			jobDisplay = " Administrator"
		} else if character.Level > LevelHero {
			jobDisplay = "   Immortal   "
		} else if character.Level == LevelHero {
			jobDisplay = "     Hero     "
		}

		var locationString string = ""
		var extrasString strings.Builder

		if character.Room != nil {
			/* Inherit the zone's who tag if we are in a room at all */
			if character.Room.Zone != nil {
				locationString = character.Room.Zone.WhoDescription
			}

			/* Hardcoded locations */
			if character.Room.Id == RoomLimbo {
				locationString = "Limbo"
			} else if character.Room.Id == RoomDeveloperLounge {
				locationString = "Office"
			}

			if character.Room.Flags&ROOM_DUNGEON != 0 && character.Room.Cell != nil {
				locationString = "Dungeon"
			}
		}

		if character.Fighting != nil {
			extrasString.WriteString("{M[<FIGHTING>]{x ")
		}

		if character.Level >= LevelHero {
			buf.WriteString(fmt.Sprintf("[%-15s][%-7s] %s %s(%s) %s\r\n",
				jobDisplay,
				locationString,
				character.Name,
				flagsString.String(),
				character.Race.DisplayName,
				extrasString.String()))
		} else {
			buf.WriteString(fmt.Sprintf("[%3d][%-10s][%-7s] %s %s(%s) %s\r\n",
				character.Level,
				jobDisplay,
				locationString,
				character.Name,
				flagsString.String(),
				character.Race.DisplayName,
				extrasString.String()))
		}
	}

	buf.WriteString(fmt.Sprintf("\r\n%d players online.\r\n", len(characters)))
	ch.Send(buf.String())
}

func do_time(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("{GThe current server time is: {g%s\r\n", time.Now().Format(time.RFC1123)))
	buf.WriteString(fmt.Sprintf("{YServer has been up since:   {y%s{x\r\n", ch.Game.startedAt.Format(time.RFC1123)))

	ch.Send(buf.String())
}

func do_look(ch *Character, arguments string) {
	var buf strings.Builder
	var obj *ObjectInstance = nil

	if ch.Room == nil {
		ch.Send("{DYou look around in the void.  There's nothing here, yet!{x\r\n")
		return
	}

	if len(arguments) > 0 {
		var found *ObjectInstance = ch.findObjectOnSelf(arguments)
		if found != nil {
			obj = found
		}

		found = ch.findObjectInRoom(arguments)
		if found != nil {
			obj = found
		}

		if obj != nil {
			ch.examineObject(obj)
			return
		}

		var foundCh *Character = ch.FindCharacterInRoom(arguments)
		if foundCh != nil {
			if foundCh != ch {
				ch.Send(fmt.Sprintf("{GYou look at %s{G.{x\r\n", foundCh.GetShortDescription(ch)))
				foundCh.Send(fmt.Sprintf("{G%s{G looks at you.{x\r\n", ch.GetShortDescriptionUpper(foundCh)))

				for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
					rch := iter.Value.(*Character)

					if rch != ch && rch != foundCh {
						rch.Send(fmt.Sprintf("{G%s{G looks at %s{G.{x\r\n", ch.GetShortDescriptionUpper(rch), foundCh.GetShortDescription(rch)))
					}
				}
			}

			ch.examineCharacter(foundCh)
			return
		}
	}

	var lookCompassOutput map[uint]string = make(map[uint]string)
	for k := uint(0); k < DirectionMax; k++ {
		if ch.Room.Exit[k] != nil {
			if ch.Room.Exit[k].Flags&EXIT_CLOSED != 0 && ch.Room.Exit[k].Flags&EXIT_LOCKED != 0 {
				lookCompassOutput[k] = "{R#"
			} else if ch.Room.Exit[k].Flags&EXIT_CLOSED != 0 {
				lookCompassOutput[k] = "{M#"
			} else {
				lookCompassOutput[k] = fmt.Sprintf("{Y%s", ExitCompassName[k])
			}
		} else {
			lookCompassOutput[k] = "{D-"
		}
	}

	var roomFlagDescriptionColour string = ""
	var roomFlagDescription string = ""

	if ch.Room.Flags&ROOM_SAFE != 0 {
		roomFlagDescriptionColour = "{W"
		roomFlagDescription = "This is a sanctuary."
	}

	buf.WriteString(fmt.Sprintf("\r\n{Y  %-50s {D-      %s{D      -\r\n", ch.Room.Name, lookCompassOutput[DirectionNorth]))
	buf.WriteString(fmt.Sprintf("{D(--------------------------------------------------) %s{D <-%s{D-{w({W*{w){D-%s{D-> %s\r\n", lookCompassOutput[DirectionWest], lookCompassOutput[DirectionUp], lookCompassOutput[DirectionDown], lookCompassOutput[DirectionEast]))
	buf.WriteString(fmt.Sprintf("{D  %s%-50s {D-      %s{D      -\r\n", roomFlagDescriptionColour, roomFlagDescription, lookCompassOutput[DirectionSouth]))
	buf.WriteString(fmt.Sprintf("\r\n{w  %s{x\r\n", ch.Room.Description))

	if len(ch.Room.Exit) > 0 {
		var exitsString strings.Builder

		for direction := uint(0); direction < DirectionMax; direction++ {
			_, ok := ch.Room.Exit[direction]
			if ok {
				if ch.Room.Exit[direction].Flags&EXIT_CLOSED != 0 {
					exitsString.WriteString("#")
				}

				exitsString.WriteString(fmt.Sprintf("%s ", ExitName[direction]))
			}
		}

		buf.WriteString(fmt.Sprintf("\r\n{W[Exits: %s]{x\r\n", strings.TrimRight(exitsString.String(), " ")))
	}

	ch.Send(buf.String())

	ch.Room.listObjectsToCharacter(ch)
	ch.Room.listOtherRoomCharactersToCharacter(ch)
}
