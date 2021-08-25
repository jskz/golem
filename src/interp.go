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
	CommandTable["say"] = Command{Name: "say", CmdFunc: do_say}
	CommandTable["save"] = Command{Name: "save", CmdFunc: do_save}

	/* act_info.go */
	CommandTable["help"] = Command{Name: "help", CmdFunc: do_help}
	CommandTable["look"] = Command{Name: "look", CmdFunc: do_look}
	CommandTable["quit"] = Command{Name: "quit", CmdFunc: do_quit}
	CommandTable["score"] = Command{Name: "score", CmdFunc: do_score}
	CommandTable["who"] = Command{Name: "who", CmdFunc: do_who}
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
		/* Send a no such command if there was any command text */
		if len(command) > 0 {
			ch.send(fmt.Sprintf("{RAlas, there is no such command: %s{x\r\n", command))
		} else {
			/* We'll still want a prompt on no input */
			ch.send("\r\n")
		}
		return false
	}

	/* Call the command func with the remaining command words joined. */
	val.CmdFunc(ch, strings.Join(words, " "))
	return true
}
