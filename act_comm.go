package main

import (
	"fmt"
	"strings"
)

func do_ooc(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString(fmt.Sprintf("[OOC] %s: %s", ch.name, arguments))
	output := buf.String()

	for client := range ch.client.game.clients {
		if client.character != nil && client.connectionState == ConnectionStatePlaying {
			client.character.send(output)
		}
	}
}
