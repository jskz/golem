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

	"github.com/dop251/goja"
)

type Command struct {
	Name         string
	MinimumLevel uint
	CmdFunc      func(ch *Character, arguments string)
	Scripted     bool
	Callback     goja.Callable
	Hidden       bool
}

var CommandTable map[string]Command

/* Magic method will be called automatically to populate command table global */
func init() {
	CommandTable = make(map[string]Command)

	/* Commands table entries which are manually initialized, grouped by file */

	/* act_comm.go */
	CommandTable["afk"] = Command{Name: "afk", CmdFunc: do_afk}
	CommandTable["ooc"] = Command{Name: "ooc", CmdFunc: do_ooc}
	CommandTable["say"] = Command{Name: "say", CmdFunc: do_say}
	CommandTable["save"] = Command{Name: "save", CmdFunc: do_save}

	/* act_info.go */
	CommandTable["help"] = Command{Name: "help", CmdFunc: do_help}
	CommandTable["look"] = Command{Name: "look", CmdFunc: do_look}
	CommandTable["quit"] = Command{Name: "quit", CmdFunc: do_quit}
	CommandTable["score"] = Command{Name: "score", CmdFunc: do_score}
	CommandTable["who"] = Command{Name: "who", CmdFunc: do_who}

	/* act_move.go */
	CommandTable["north"] = Command{Name: "north", CmdFunc: do_north}
	CommandTable["east"] = Command{Name: "east", CmdFunc: do_east}
	CommandTable["south"] = Command{Name: "south", CmdFunc: do_south}
	CommandTable["west"] = Command{Name: "west", CmdFunc: do_west}
	CommandTable["up"] = Command{Name: "up", CmdFunc: do_up}
	CommandTable["down"] = Command{Name: "down", CmdFunc: do_down}
	CommandTable["follow"] = Command{Name: "follow", CmdFunc: do_follow}

	/* act_obj.go */
	CommandTable["equipment"] = Command{Name: "equipment", CmdFunc: do_equipment}
	CommandTable["inventory"] = Command{Name: "inventory", CmdFunc: do_inventory}
	CommandTable["wear"] = Command{Name: "wear", CmdFunc: do_wear}
	CommandTable["remove"] = Command{Name: "remove", CmdFunc: do_remove}
	CommandTable["take"] = Command{Name: "take", CmdFunc: do_take}
	CommandTable["drop"] = Command{Name: "drop", CmdFunc: do_drop}

	/* act_wiz.go */
	CommandTable["exec"] = Command{Name: "exec", CmdFunc: do_exec, MinimumLevel: LevelAdmin}
	CommandTable["goto"] = Command{Name: "goto", CmdFunc: do_goto, MinimumLevel: LevelHero + 1}
	CommandTable["mem"] = Command{Name: "mem", CmdFunc: do_mem, MinimumLevel: LevelAdmin}
	CommandTable["mlist"] = Command{Name: "mlist", CmdFunc: do_mlist, MinimumLevel: LevelAdmin}
	CommandTable["path"] = Command{Name: "path", CmdFunc: do_path, MinimumLevel: LevelAdmin}
	CommandTable["peace"] = Command{Name: "peace", CmdFunc: do_peace, MinimumLevel: LevelHero + 1}
	CommandTable["purge"] = Command{Name: "purge", CmdFunc: do_purge, MinimumLevel: LevelHero + 2}
	CommandTable["shutdown"] = Command{Name: "shutdown", CmdFunc: do_shutdown, MinimumLevel: LevelAdmin}
	CommandTable["zones"] = Command{Name: "zones", CmdFunc: do_zones, MinimumLevel: LevelHero + 1}

	/* fight.go */
	CommandTable["flee"] = Command{Name: "flee", CmdFunc: do_flee}
	CommandTable["kill"] = Command{Name: "kill", CmdFunc: do_kill}

	/* magic.go */
	CommandTable["cast"] = Command{Name: "cast", CmdFunc: do_cast}
	CommandTable["spells"] = Command{Name: "spells", CmdFunc: do_spells}

	/* scripting.go */
	CommandTable["reload"] = Command{Name: "reload", CmdFunc: do_reload, MinimumLevel: LevelAdmin}

	/* skills.go */
	CommandTable["practice"] = Command{Name: "practice", CmdFunc: do_practice}
	CommandTable["skills"] = Command{Name: "skills", CmdFunc: do_skills}

	/* Aliases */
	CommandTable["eq"] = Command{Name: "equipment", CmdFunc: do_equipment, Hidden: true}
	CommandTable["i"] = Command{Name: "inventory", CmdFunc: do_inventory, Hidden: true}
	CommandTable["k"] = Command{Name: "kill", CmdFunc: do_kill, Hidden: true}
	CommandTable["l"] = Command{Name: "look", CmdFunc: do_look, Hidden: true}
	CommandTable["get"] = Command{Name: "take", CmdFunc: do_take, Hidden: true}
	CommandTable["n"] = Command{Name: "north", CmdFunc: do_north, Hidden: true}
	CommandTable["e"] = Command{Name: "east", CmdFunc: do_east, Hidden: true}
	CommandTable["s"] = Command{Name: "south", CmdFunc: do_south, Hidden: true}
	CommandTable["w"] = Command{Name: "west", CmdFunc: do_west, Hidden: true}
	CommandTable["u"] = Command{Name: "up", CmdFunc: do_up, Hidden: true}
	CommandTable["d"] = Command{Name: "down", CmdFunc: do_down, Hidden: true}
}

func (ch *Character) Interpret(input string) bool {
	defer func() {
		recover()
	}()

	words := strings.Split(input, " ")
	if len(words) < 1 {
		return false
	}

	/* Extract the command and shift it out of the input words */
	command, words := strings.ToLower(words[0]), words[1:]

	val, ok := CommandTable[command]
	if !ok || (ok && ch.level < val.MinimumLevel) {
		/* Send a no such command if there was any command text */
		if len(command) > 0 {
			ch.Send(fmt.Sprintf("{RAlas, there is no such command: %s{x\r\n", command))
		} else {
			/* We'll still want a prompt on no input */
			ch.Send("\r\n")
			return true
		}

		return false
	}

	rest := strings.Join(words, " ")

	/* Call the command func with the remaining command words joined. */
	if val.Scripted {
		val.Callback(ch.game.vm.ToValue(ch), ch.game.vm.ToValue(ch), ch.game.vm.ToValue(rest))
		return true
	}

	val.CmdFunc(ch, rest)
	return true
}
