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
		commands = append(commands, command.Name)
	}

	sort.Strings(commands)

	for _, command := range commands {
		buf.WriteString(fmt.Sprintf("%-10s ", command))
		index++

		if index%7 == 0 {
			buf.WriteString("\r\n")
		}
	}

	if index%7 != 0 {
		buf.WriteString("\r\n")
	}

	ch.send(buf.String())
}

/* Display relevant game information about the player's character. */
func do_score(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("{D┌─ {WCharacter Information {D────────┐{x\r\n")

	buf.WriteString(fmt.Sprintf("{D│ {wName: %-16s         {D│\r\n", ch.name))
	buf.WriteString(fmt.Sprintf("{D│ {wLevel: %-3d                     {D│\r\n", ch.level))
	buf.WriteString(fmt.Sprintf("{D│ {wRace: %-17s        {D│\r\n", RaceTable[ch.race].DisplayName))
	buf.WriteString(fmt.Sprintf("{D│ {wJob: %-17s         {D│\r\n", JobsTable[ch.job].DisplayName))
	buf.WriteString(fmt.Sprintf("{D│ {wHealth: %-11s            {D│\r\n", fmt.Sprintf("%d/%d", ch.health, ch.maxHealth)))
	buf.WriteString(fmt.Sprintf("{D│ {wMana: %-11s              {D│\r\n", fmt.Sprintf("%d/%d", ch.mana, ch.maxMana)))
	buf.WriteString(fmt.Sprintf("{D│ {wStamina: %-11s           {D│\r\n", fmt.Sprintf("%d/%d", ch.stamina, ch.maxStamina)))
	buf.WriteString("{D└────────────────────────────────┘{x\r\n")

	output := buf.String()
	ch.send(output)
}

/* Display a list of players online (ad visible to the current player character!) */
func do_who(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("{CThe following players are online:{x\r\n")

	for client := range ch.client.game.clients {
		/* If the client is "at least" playing, then we will display them in the WHO list */
		if client.connectionState >= ConnectionStatePlaying && client.character != nil {
			buf.WriteString(fmt.Sprintf("[%-10s] %s (%s)\r\n",
				JobsTable[client.character.job].DisplayName,
				client.character.name,
				RaceTable[client.character.race].DisplayName))
		}
	}

	output := buf.String()
	ch.send(output)
}

func do_look(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("You look around into the void.\r\n")
	output := buf.String()
	ch.send(output)
}
