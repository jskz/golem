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
	"sort"
	"strings"
)

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
		if ch.level < CommandTable[command].MinimumLevel || CommandTable[command].Hidden {
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

	healthPercentage := ch.health * 100 / ch.maxHealth
	manaPercentage := ch.mana * 100 / ch.maxMana
	staminaPercentage := ch.stamina * 100 / ch.maxStamina

	currentHealthColour := SeverityColourFromPercentage(healthPercentage)
	currentManaColour := SeverityColourFromPercentage(manaPercentage)
	currentStaminaColour := SeverityColourFromPercentage(staminaPercentage)

	buf.WriteString("\r\n{D┌─ {WCharacter Information {D──────────────────┬─ {WStatistics{D ───────┐{x\r\n")
	buf.WriteString(fmt.Sprintf("{D│ {CName:    {c%-13s                   {D│ Strength:       {M%2d{D │\r\n", ch.name, ch.strength))
	if ch.level < LevelHero {
		buf.WriteString(fmt.Sprintf("{D│ {CLevel:   {c%-3d  {D[%8d exp. until next] {D│ Dexterity:      {M%2d{D │\r\n", ch.level, ExperienceRequiredForLevel(int(ch.level+1))-int(ch.experience), ch.dexterity))
	} else {
		buf.WriteString(fmt.Sprintf("{D│ {CLevel:   {c%-3d                             {D│ Dexterity:      {M%2d{D │\r\n", ch.level, ch.dexterity))
	}
	buf.WriteString(fmt.Sprintf("{D│ {CRace:    {c%-21s           {D│ Intelligence:   {M%2d{D │\r\n", ch.race.DisplayName, ch.intelligence))
	buf.WriteString(fmt.Sprintf("{D│ {CJob:     {c%-21s           {D│ Wisdom:         {M%2d{D │\r\n", ch.job.DisplayName, ch.wisdom))
	buf.WriteString(fmt.Sprintf("{D│ {CHealth:  {c%s%-20s                {D│ Constitution:   {M%2d{D │\r\n",
		currentHealthColour,
		fmt.Sprintf("%-5d{w/{G%-5d",
			ch.health,
			ch.maxHealth),
		ch.constitution))
	buf.WriteString(fmt.Sprintf("{D│ {CMana:    {c%s%-18s                  {D│ Charisma:       {M%2d{D │\r\n",
		currentManaColour,
		fmt.Sprintf("%-5d{w/{G%-5d",
			ch.mana,
			ch.maxMana),
		ch.charisma))
	buf.WriteString(fmt.Sprintf("{D│ {CStamina: {c%s%-21s               {D│ Luck:           {M%2d{D │\r\n",
		currentStaminaColour,
		fmt.Sprintf("%-5d{w/{G%-5d",
			ch.stamina,
			ch.maxStamina),
		ch.luck))
	buf.WriteString("{D└──────────────────────────────────────────┴────────────────────┘{x\r\n")

	output := buf.String()
	ch.Send(output)
}

/* Display a list of players online (and visible to the current player character!) */
func do_who(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("\r\n{CThe following players are online:{x\r\n")

	characters := make([]*Character, 0)
	for client := range ch.game.clients {
		if client.character != nil && client.connectionState >= ConnectionStatePlaying {
			characters = append(characters, client.character)
		}
	}

	sort.Slice(characters, func(i int, j int) bool {
		return characters[i].level > characters[j].level
	})

	for _, character := range characters {
		var flagsString strings.Builder

		if character.afk != nil {
			flagsString.WriteString("{G[AFK]{x ")
		}

		jobDisplay := character.job.DisplayName
		if character.level == LevelAdmin {
			jobDisplay = " Administrator"
		} else if character.level > LevelHero {
			jobDisplay = "   Immortal   "
		} else if character.level == LevelHero {
			jobDisplay = "     Hero     "
		}

		if character.level >= LevelHero {
			buf.WriteString(fmt.Sprintf("[%-15s] %s %s(%s)\r\n",
				jobDisplay,
				character.name,
				flagsString.String(),
				character.race.DisplayName))
		} else {
			buf.WriteString(fmt.Sprintf("[%3d][%-10s] %s %s(%s)\r\n",
				character.level,
				jobDisplay,
				character.name,
				flagsString.String(),
				character.race.DisplayName))
		}
	}

	buf.WriteString(fmt.Sprintf("\r\n%d players online.\r\n", len(characters)))
	ch.Send(buf.String())
}

func do_look(ch *Character, arguments string) {
	var buf strings.Builder

	if ch.room == nil {
		ch.Send("{DYou look around into the void.  There's nothing here, yet!{x\r\n")
		return
	}

	var lookCompassOutput map[uint]string = make(map[uint]string)
	for k := uint(0); k < DirectionMax; k++ {
		if ch.room.exit[k] != nil {
			lookCompassOutput[k] = fmt.Sprintf("{Y%s", ExitCompassName[k])
		} else {
			lookCompassOutput[k] = "{D-"
		}
	}

	buf.WriteString(fmt.Sprintf("\r\n{Y  %-50s {D-      %s{D      -\r\n", ch.room.name, lookCompassOutput[DirectionNorth]))
	buf.WriteString(fmt.Sprintf("{D(--------------------------------------------------) %s{D <-%s{D-{w({W*{w){D-%s{D-> %s\r\n", lookCompassOutput[DirectionWest], lookCompassOutput[DirectionUp], lookCompassOutput[DirectionDown], lookCompassOutput[DirectionEast]))
	buf.WriteString(fmt.Sprintf("{D                                                     {D-      %s{D      -\r\n", lookCompassOutput[DirectionSouth]))
	buf.WriteString(fmt.Sprintf("{w  %s{x\r\n", ch.room.description))

	if len(ch.room.exit) > 0 {
		var exitsString strings.Builder

		for direction := uint(0); direction < DirectionMax; direction++ {
			_, ok := ch.room.exit[direction]
			if ok {
				exitsString.WriteString(ExitName[direction])
			}
		}

		buf.WriteString(fmt.Sprintf("\r\n{W[Exits: %s]{x\r\n", exitsString.String()))
	}

	ch.Send(buf.String())

	ch.room.listObjectsToCharacter(ch)
	ch.room.listOtherRoomCharactersToCharacter(ch)
}
