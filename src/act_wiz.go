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
	"strconv"
	"strings"
	"time"
)

func (ch *Character) isAdmin() bool {
	return ch.Level == LevelAdmin
}

func do_exec(ch *Character, arguments string) {
	if ch.Client != nil {
		value, err := ch.Game.vm.RunString(arguments)

		if err != nil {
			ch.Send(fmt.Sprintf("{R\r\nError: %s{x.\r\n", err.Error()))
			return
		}

		ch.Send(fmt.Sprintf("{w\r\n%s{x\r\n", value.String()))
	}
}

func do_zones(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("{Y%-4s %-35s [%-11s] %s/%s\r\n",
		"ID#",
		"Zone Name",
		"Low# -High#",
		"Reset Freq.",
		"Min. Since"))

	for iter := ch.Game.Zones.Head; iter != nil; iter = iter.Next {
		zone := iter.Value.(*Zone)

		minutesSinceZoneReset := int(time.Since(zone.LastReset).Minutes())

		output.WriteString(fmt.Sprintf("%-4d %-35s [%-5d-%-5d] %d/%d\r\n",
			zone.Id,
			zone.Name,
			zone.Low,
			zone.High,
			zone.ResetFrequency,
			minutesSinceZoneReset))
	}

	output.WriteString("{x")
	ch.Send(output.String())
}

func do_mem(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString("{YUsage statistics:\r\n")
	output.WriteString(fmt.Sprintf("%-15s %-6d\r\n", "Characters", ch.Game.Characters.Count))
	output.WriteString(fmt.Sprintf("%-15s %-6d\r\n", "Jobs", Jobs.Count))
	output.WriteString(fmt.Sprintf("%-15s %-6d\r\n", "Races", Races.Count))
	output.WriteString(fmt.Sprintf("%-15s %-6d{x\r\n", "Zones", ch.Game.Zones.Count))

	ch.Send(output.String())
}

func do_mlist(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Displaying all %d character instances in the world:\r\n", ch.Game.Characters.Count))

	for iter := ch.Game.Characters.Head; iter != nil; iter = iter.Next {
		wch := iter.Value.(*Character)

		if wch.Flags&CHAR_IS_PLAYER != 0 {
			if wch.Client != nil {
				output.WriteString(fmt.Sprintf("{G%s@%s{x\r\n", wch.Name, wch.Client.conn.RemoteAddr().String()))
			} else {
				output.WriteString(fmt.Sprintf("{G%s@DISCONNECTED{x\r\n", wch.Name))
			}
		} else {
			output.WriteString(fmt.Sprintf("%s{x (id#%d)\r\n", wch.GetShortDescriptionUpper(ch), wch.Id))
		}
	}

	ch.Send(output.String())
}

func do_purge(ch *Character, arguments string) {
	if ch.Room == nil {
		return
	}

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)
		if rch == ch || rch.Client != nil || rch.Flags&CHAR_IS_PLAYER != 0 {
			continue
		}

		ch.Room.Characters.Remove(rch)
	}

	for {
		if ch.Room.Objects.Head == nil {
			break
		}

		ch.Room.Objects.Remove(ch.Room.Objects.Head.Value)
	}

	ch.Send("You have purged the contents of the room.\r\n")

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if !rch.IsEqual(ch) {
			rch.Send(fmt.Sprintf("%s purges the contents of the room.\r\n", ch.GetShortDescriptionUpper(rch)))
		}
	}
}

func do_peace(ch *Character, arguments string) {
	if ch.Room == nil || ch.Client == nil {
		return
	}

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		rch.Flags &= ^CHAR_AGGRESSIVE
		rch.Fighting = nil
		rch.Combat = nil
	}

	ch.Send("Ok.\r\n")
}

func do_shutdown(ch *Character, arguments string) {
	if ch.Client != nil {
		ch.Game.shutdownRequest <- true
	}
}

func do_wiznet(ch *Character, arguments string) {
	if ch.Wiznet {
		ch.Wiznet = false
		ch.Send("Wiznet disabled.\r\n")
		return
	}

	ch.Wiznet = true
	ch.Send("Wiznet enabled.\r\n")
}

func do_webhook(ch *Character, arguments string) {
	if len(arguments) < 1 {
		output := "{WWebhook management:\r\n" +
			"{Glist       - {glist all system webhooks\r\n" +
			"{Gcreate     - {gcreate a system webhook\r\n" +
			"{Gdelete [#] - {gdelete a webhook by ID (from {G\"webhook list\"){x\r\n"
		ch.Send(output)
		return
	}

	firstArgument, arguments := oneArgument(arguments)

	command := strings.ToLower(firstArgument)
	switch command {
	case "list":
		var output strings.Builder

		output.WriteString("{Y  ID# | URL\r\n")
		output.WriteString("------+------------------------------------------------------------------------\r\n")

		for _, webhook := range ch.Game.webhooks {
			output.WriteString(fmt.Sprintf("{Y%5d | %swebhook?key=%s\r\n", webhook.Id, Config.WebConfiguration.PublicRoot, webhook.Uuid))
		}

		output.WriteString("{x")
		ch.Send(output.String())

	case "create":
		webhook, err := ch.Game.CreateWebhook()
		if err != nil {
			ch.Send(fmt.Sprintf("Something went wrong trying to create a new webhook: %v\r\n", err))
			break
		}

		ch.Send(fmt.Sprintf("Successfully created a new webhook with URL:\r\n{Y%swebhook?key=%s{x\r\n", Config.WebConfiguration.PublicRoot, webhook.Uuid))

	case "delete":
		secondArgument, _ := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Delete requires an ID argument.\r\n")
			break
		}

		id, err := strconv.Atoi(secondArgument)
		if err != nil {
			ch.Send("Bad argument, please provider an integer ID.\r\n")
			break
		}

		for _, webhook := range ch.Game.webhooks {
			if webhook.Id == id {
				err := ch.Game.DeleteWebhook(webhook)
				if err != nil {
					ch.Send(fmt.Sprintf("Something went wrong trying to delete that webhook: %v\r\n", err))
					return
				}

				ch.Send("Ok.\r\n")
				return
			}
		}

		ch.Send("A webook with that ID could not be found.\r\n")

	default:
		ch.Send("Unrecognized command.\r\n")
	}
}

func do_script(ch *Character, arguments string) {
	if len(arguments) < 1 {
		output := "{WScript management:\r\n" +
			"{Glist       - {glist all mutable scripts\r\n" +
			"{Gcreate     - {gcreate a new mutable script\r\n" +
			"{Gshow [#]   - {gshow more details about a mutable script\r\n" +
			"{Gedit [#]   - {gstart a line editor on a script's source\r\n" +
			"{Gdelete [#] - {gdelete a script by ID (from {G\"script list\"){x\r\n"
		ch.Send(output)
		return
	}

	firstArgument, arguments := oneArgument(arguments)

	command := strings.ToLower(firstArgument)
	switch command {
	case "list":
		var output strings.Builder

		output.WriteString("{Y  ID# | Name\r\n")
		output.WriteString("------+-------------------------------------\r\n")

		for _, script := range ch.Game.scripts {
			output.WriteString(fmt.Sprintf("{Y%5d | %s\r\n", script.id, script.name))
		}

		output.WriteString("{x")
		ch.Send(output.String())

	case "create":
		secondArgument, _ := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Create requires a script name string argument.\r\n")
			break
		}

		script, err := ch.Game.CreateScript(secondArgument, "module.exports = {};")
		if err != nil {
			ch.Send(fmt.Sprintf("Something went wrong trying to create a new script: %v\r\n", err))
			break
		}

		ch.Send(fmt.Sprintf("Successfully created a new script with ID %d.{x\r\n", script.id))

	case "delete":
		secondArgument, _ := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Delete requires an ID argument.\r\n")
			break
		}

		id, err := strconv.Atoi(secondArgument)
		if err != nil {
			ch.Send("Bad argument, please provider an integer ID.\r\n")
			break
		}

		for _, script := range ch.Game.scripts {
			if script.id == uint(id) {
				err := ch.Game.DeleteScript(script)
				if err != nil {
					ch.Send(fmt.Sprintf("Something went wrong trying to delete this script: %v\r\n", err))
					return
				}

				ch.Send("Ok.\r\n")
				return
			}
		}

		ch.Send("A script with that ID could not be found.\r\n")

	default:
		ch.Send("Unrecognized command.\r\n")
	}
}

func do_goto(ch *Character, arguments string) {
	id, err := strconv.Atoi(arguments)
	if err != nil || id <= 0 {
		ch.Send("Goto which room ID?\r\n")
		return
	}

	room, err := ch.Game.LoadRoomIndex(uint(id))
	if err != nil || room == nil {
		ch.Send("No such room.\r\n")
		return
	}

	if ch.Room != nil {
		for iter := room.Characters.Head; iter != nil; iter = iter.Next {
			character := iter.Value.(*Character)
			if character != ch {
				character.Send(fmt.Sprintf("\r\n{W%s{W disappears in a puff of smoke.{x\r\n", ch.GetShortDescriptionUpper(character)))
			}
		}

		ch.Room.removeCharacter(ch)
	}

	room.AddCharacter(ch)

	for iter := room.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		if character != ch {
			character.Send(fmt.Sprintf("\r\n{W%s{W appears in a puff of smoke.{x\r\n", ch.GetShortDescriptionUpper(character)))
		}
	}

	do_look(ch, "")
}
