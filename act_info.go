package main

import (
	"strings"
)

func do_score(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("Character information:\r\n")
	output := buf.String()
	ch.send(output)
}

func do_look(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("You look around into the void.\r\n")
	output := buf.String()
	ch.send(output)
}
