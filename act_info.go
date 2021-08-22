package main

import (
	"fmt"
	"strings"
)

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
	output := buf.String()
	ch.send(output)
}

func do_look(ch *Character, arguments string) {
	var buf strings.Builder

	buf.WriteString("You look around into the void.\r\n")
	output := buf.String()
	ch.send(output)
}
