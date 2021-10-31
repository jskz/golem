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
	CommandTable["group"] = Command{Name: "group", CmdFunc: do_group}
	CommandTable["ooc"] = Command{Name: "ooc", CmdFunc: do_ooc}
	CommandTable["say"] = Command{Name: "say", CmdFunc: do_say}
	CommandTable["save"] = Command{Name: "save", CmdFunc: do_save}

	/* act_info.go */
	CommandTable["help"] = Command{Name: "help", CmdFunc: do_help}
	CommandTable["look"] = Command{Name: "look", CmdFunc: do_look}
	CommandTable["quit"] = Command{Name: "quit", CmdFunc: do_quit}
	CommandTable["score"] = Command{Name: "score", CmdFunc: do_score}
	CommandTable["who"] = Command{Name: "who", CmdFunc: do_who}
	CommandTable["time"] = Command{Name: "time", CmdFunc: do_time}

	/* act_move.go */
	CommandTable["north"] = Command{Name: "north", CmdFunc: do_north}
	CommandTable["east"] = Command{Name: "east", CmdFunc: do_east}
	CommandTable["south"] = Command{Name: "south", CmdFunc: do_south}
	CommandTable["west"] = Command{Name: "west", CmdFunc: do_west}
	CommandTable["up"] = Command{Name: "up", CmdFunc: do_up}
	CommandTable["down"] = Command{Name: "down", CmdFunc: do_down}
	CommandTable["follow"] = Command{Name: "follow", CmdFunc: do_follow}
	CommandTable["open"] = Command{Name: "open", CmdFunc: do_open}
	CommandTable["close"] = Command{Name: "close", CmdFunc: do_close}

	/* act_obj.go */
	CommandTable["equipment"] = Command{Name: "equipment", CmdFunc: do_equipment}
	CommandTable["inventory"] = Command{Name: "inventory", CmdFunc: do_inventory}
	CommandTable["wear"] = Command{Name: "wear", CmdFunc: do_wear}
	CommandTable["remove"] = Command{Name: "remove", CmdFunc: do_remove}
	CommandTable["give"] = Command{Name: "give", CmdFunc: do_give}
	CommandTable["take"] = Command{Name: "take", CmdFunc: do_take}
	CommandTable["drop"] = Command{Name: "drop", CmdFunc: do_drop}
	CommandTable["use"] = Command{Name: "use", CmdFunc: do_use}

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
	CommandTable["wiznet"] = Command{Name: "wiznet", CmdFunc: do_wiznet, MinimumLevel: LevelAdmin}

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
	if ch.outputCursor > 0 && ch.inputCursor < ch.outputHead {
		/* If any input, abort the paging */
		if input != "" {
			ch.clearOutputBuffer()
			ch.Client.displayPrompt()
			return true
		}

		ch.inputCursor += DefaultMaxLines
		return true
	}

	if ch.Client != nil && ch.Client.ConnectionHandler != nil {
		(*ch.Client.ConnectionHandler)(ch.Game.vm.ToValue(ch.Client), ch.Game.vm.ToValue(input))
		return true
	}

	words := strings.Split(input, " ")
	if len(words) < 1 {
		return false
	}

	/* Extract the command and shift it out of the input words */
	command, words := strings.ToLower(words[0]), words[1:]
	rest := strings.TrimSpace(strings.Join(words, " "))

	val, ok := CommandTable[command]
	if !ok || (ok && ch.Level < val.MinimumLevel) {
		/* Send a no such command if there was any command text */
		if len(command) > 0 {
			/* As a fallback, see if this command matches any proficiency which has a registered handler. */
			prof := ch.FindProficiencyByName(command)
			if prof == nil || prof.Proficiency <= 0 || ch.Game.skills[prof.SkillId].Handler == nil {
				ch.Send(fmt.Sprintf("{RAlas, there is no such command: %s{x\r\n", command))
				return false
			}

			if ch.Game.skills[prof.SkillId].Intent == SkillIntentOffensive && ch.Room.Flags&ROOM_SAFE != 0 {
				ch.Send("You can't do that here.\r\n")
				return false
			}

			(*ch.Game.skills[prof.SkillId].Handler)(ch.Game.vm.ToValue(prof), ch.Game.vm.ToValue(ch), ch.Game.vm.ToValue(rest))
		} else {
			/* We'll still want a prompt on no input */
			ch.Send("\r\n")
		}

		return true
	}
	/* Call the command func with the remaining command words joined. */
	if val.Scripted {
		val.Callback(ch.Game.vm.ToValue(ch), ch.Game.vm.ToValue(ch), ch.Game.vm.ToValue(rest))
		return true
	}

	val.CmdFunc(ch, rest)
	return true
}
