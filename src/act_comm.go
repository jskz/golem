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

func do_ooc(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("[OOC] %s: %s{x", ch.name, arguments))
	output := buf.String()

	for client := range ch.client.game.clients {
		if client.character != nil && client.connectionState == ConnectionStatePlaying {
			client.character.send(output)
		}
	}
}
