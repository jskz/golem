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

	buf.WriteString("\r\n{D┌─ {WCharacter Information {D──────────────────┐{x\r\n")
	buf.WriteString(fmt.Sprintf("{D│ {wName: %-16s                   {D│\r\n", ch.name))
	buf.WriteString(fmt.Sprintf("{D│ {wLevel: %-3d                               {D│\r\n", ch.level))
	if ch.level < LevelHero {
		buf.WriteString(fmt.Sprintf("{D│ {wExperience: %-7d (%-7d until next) {D│\r\n", ch.experience, ExperienceRequiredForLevel(int(ch.level+1))-int(ch.experience)))
	}
	buf.WriteString(fmt.Sprintf("{D│ {wRace: %-21s              {D│\r\n", ch.race.DisplayName))
	buf.WriteString(fmt.Sprintf("{D│ {wJob: %-21s               {D│\r\n", ch.job.DisplayName))
	buf.WriteString(fmt.Sprintf("{D│ {wHealth: %-15s                  {D│\r\n", fmt.Sprintf("%d/%d", ch.health, ch.maxHealth)))
	buf.WriteString(fmt.Sprintf("{D│ {wMana: %-15s                    {D│\r\n", fmt.Sprintf("%d/%d", ch.mana, ch.maxMana)))
	buf.WriteString(fmt.Sprintf("{D│ {wStamina: %-15s                 {D│\r\n", fmt.Sprintf("%d/%d", ch.stamina, ch.maxStamina)))
	buf.WriteString("{D└──────────────────────────────────────────┘{x\r\n")

	output := buf.String()
	ch.Send(output)
}

/* Display a list of players online (and visible to the current player character!) */
func do_who(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("\r\n{CThe following players are online:{x\r\n")

	characters := make([]*Character, 0)
	for client := range ch.client.game.clients {
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

		buf.WriteString(fmt.Sprintf("[%3d][%-10s] %s %s(%s)\r\n",
			character.level,
			character.job.DisplayName,
			character.name,
			flagsString.String(),
			character.race.DisplayName))
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
	buf.WriteString(fmt.Sprintf("{w   %s{x\r\n", ch.room.description))

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
