package main

import (
	"fmt"
	"strings"
)

type Command struct {
	Name    string
	CmdFunc func(ch *Character, arguments string)
}

var CommandTable map[string]Command

/* Magic method will be called automatically to populate command table global */
func init() {
	CommandTable = make(map[string]Command)

	/* Commands table entries which are manually initialized, grouped by file */

	/* act_comm.go */
	CommandTable["ooc"] = Command{Name: "ooc", CmdFunc: do_ooc}

	/* act_info.go */
	CommandTable["look"] = Command{Name: "look", CmdFunc: do_look}
	CommandTable["score"] = Command{Name: "score", CmdFunc: do_score}
}

func (ch *Character) Interpret(input string) bool {
	words := strings.Split(input, " ")
	if len(words) < 1 {
		return false
	}

	/* Extract the command and shift it out of the input words */
	command, words := strings.ToLower(words[0]), words[1:]
	val, ok := CommandTable[command]
	if !ok {
		/* No such command */
		ch.send(fmt.Sprintf("Alas, there is no such command: %s.\r\n", command))
		return false
	}

	/* Call the command func with the remaining command words joined. */
	val.CmdFunc(ch, strings.Join(words, " "))
	return true
}
