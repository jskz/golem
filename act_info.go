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
	"strings"
)

/* List all commands available to the player in rows of 7 items. */
func do_help(ch *Character, arguments string) {
	var buf strings.Builder
	var index int = 0

	for _, command := range CommandTable {
		buf.WriteString(fmt.Sprintf("%-14s ", command.Name))
		index++

		if index%7 == 0 {
			buf.WriteString("\r\n")
		}
	}

	ch.send(buf.String())
}

/* Display relevant game information about the player's character. */
func do_score(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("Character information:\r\n")

	buf.WriteString(fmt.Sprintf("Name: %s\r\n", ch.name))
	buf.WriteString(fmt.Sprintf("Job: %s\r\n", JobsTable[ch.job].DisplayName))
	buf.WriteString(fmt.Sprintf("Level: %d\r\n", ch.level))

	output := buf.String()
	ch.send(output)
}

/* Display a list of players online (and visible to the current player character!) */
func do_who(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("The following players are online:\r\n")

	for client := range ch.client.game.clients {
		/* Conditionally render an admin view */

		/* If the client is "at least" playing, then we will display them in the WHO list */
		if client.connectionState >= ConnectionStatePlaying && client.character != nil {
			buf.WriteString(fmt.Sprintf("[%-10s] %s\r\n", JobsTable[client.character.job].DisplayName, client.character.name))
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
