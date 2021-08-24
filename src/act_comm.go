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

/* say will be room-specific */
func do_say(ch *Character, arguments string) {
	var buf strings.Builder

	if len(arguments) == 0 {
		ch.send("{CSay what?{x\r\n")
		return
	}

	buf.WriteString(fmt.Sprintf("{C%s says \"%s{C\"{x\r\n", ch.name, arguments))
	output := buf.String()

	ch.send(fmt.Sprintf("{CYou say \"%s{C\"{x\r\n", arguments))

	for client := range ch.client.game.clients {
		if client.character != nil && client.character != ch && client.connectionState == ConnectionStatePlaying {
			client.character.send(output)
		}
	}
}

func do_ooc(ch *Character, arguments string) {
	var buf strings.Builder

	if len(arguments) == 0 {
		ch.send("{MWhat message do you wish to globally send out-of-character?{x\r\n")
		return
	}

	buf.WriteString(fmt.Sprintf("{M[OOC] %s: %s{x\r\n", ch.name, arguments))
	output := buf.String()

	for client := range ch.client.game.clients {
		if client.character != nil && client.connectionState == ConnectionStatePlaying {
			client.character.send(output)
		}
	}
}
