package main

import (
	"fmt"
	"strings"
)

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

func do_score(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("Character information:\r\n")
	buf.WriteString(fmt.Sprintf("Name: %s\r\n", ch.name))
	buf.WriteString(fmt.Sprintf("Level: %d\r\n", ch.level))

	output := buf.String()
	ch.send(output)
}

func do_who(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("The following players are online:\r\n")

	for client := range ch.client.game.clients {
		/* If the client is "at least" playing, then we will display them in the WHO list */
		if client.connectionState >= ConnectionStatePlaying && client.character != nil {
			buf.WriteString(fmt.Sprintf("[%-7s] %s\r\n", "status", client.character.name))
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
