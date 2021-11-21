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
	output.WriteString(fmt.Sprintf("%-15s %-6d\r\n", "Objects", ch.Game.Objects.Count))
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
			"{Glist                                 - {glist all webhooks\r\n" +
			"{Gcreate                               - {gcreate webhook\r\n" +
			"{Gshow [webhook_id]                    - {gdetailed info about webhook by ID{x\r\n" +
			"{Gdelete [webhook_id]                  - {gdelete webhook by ID{x\r\n" +
			"{Gconnect [webhook_id] [script_id]     - {gconnect webhook with script{x\r\n" +
			"{Gdisconnect [webhook_id] [script_id]  - {gdetach webhook from script{x\r\n"
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

	case "disconnect":
		secondArgument, arguments := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Disconnect requires two ID arguments, webhook_id and script_id.\r\n")
			break
		}

		webhookId, err := strconv.Atoi(secondArgument)
		if err != nil {
			ch.Send("Bad argument, please provider two integer IDs.\r\n")
			break
		}

		thirdArgument, _ := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Disconnect requires two ID arguments, webhook_id and script_id.\r\n")
			break
		}

		scriptId, err := strconv.Atoi(thirdArgument)
		if err != nil {
			ch.Send("Bad argument, please provider two integer IDs.\r\n")
			break
		}

		script, ok := ch.Game.Scripts[uint(scriptId)]
		if !ok {
			ch.Send("Could not find script with that ID for webhook to disconnect.\r\n")
			break
		}

		for _, webhook := range ch.Game.webhooks {
			if webhook.Id == webhookId {
				err := webhook.DetachScript(script)
				if err != nil {
					ch.Send(fmt.Sprintf("Something went wrong trying to detach webhook-script relation: %v\r\n", err))
					return
				}

				ch.Send("Ok.\r\n")
				return
			}
		}

		ch.Send("Could not find that webhook to detach that script.\r\n")

	case "connect":
		secondArgument, arguments := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Connect requires two ID arguments, webhook_id and script_id.\r\n")
			break
		}

		webhookId, err := strconv.Atoi(secondArgument)
		if err != nil {
			ch.Send("Bad argument, please provider two integer IDs.\r\n")
			break
		}

		thirdArgument, _ := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Connect requires two ID arguments, webhook_id and script_id.\r\n")
			break
		}

		scriptId, err := strconv.Atoi(thirdArgument)
		if err != nil {
			ch.Send("Bad argument, please provider two integer IDs.\r\n")
			break
		}

		script, ok := ch.Game.Scripts[uint(scriptId)]
		if !ok {
			ch.Send("Could not find script with that ID for webhook to connect.\r\n")
			break
		}

		for _, webhook := range ch.Game.webhooks {
			if webhook.Id == webhookId {
				err := webhook.AttachScript(script)
				if err != nil {
					ch.Send(fmt.Sprintf("Something went wrong trying to connect webhook-script relation: %v\r\n", err))
					return
				}

				ch.Send("Ok.\r\n")
				return
			}
		}

		ch.Send("Could not find that webhook to attach that script.\r\n")

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

	case "show":
		secondArgument, _ := oneArgument(arguments)
		if secondArgument == "" {
			ch.Send("Show requires an ID argument.\r\n")
			break
		}

		id, err := strconv.Atoi(secondArgument)
		if err != nil {
			ch.Send("Bad argument, please provider an integer ID.\r\n")
			break
		}

		for _, webhook := range ch.Game.webhooks {
			if webhook.Id == id {
				var output strings.Builder

				output.WriteString(fmt.Sprintf("{YWebhook with ID %d:\r\n", webhook.Id))
				output.WriteString(fmt.Sprintf("URL: %swebhook?key=%s\r\n", Config.WebConfiguration.PublicRoot, webhook.Uuid))

				script, ok := ch.Game.webhookScripts[id]
				if ok {
					output.WriteString(fmt.Sprintf("{C* {MWebhook connected to script {Y%s{M/{Y%d{x\r\n", script.Name, script.Id))
				}

				ch.Send(output.String())
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

		for _, script := range ch.Game.Scripts {
			output.WriteString(fmt.Sprintf("{Y%5d | %s\r\n", script.Id, script.Name))
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

		ch.Send(fmt.Sprintf("Successfully created a new script with ID %d.{x\r\n", script.Id))

	case "edit":
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

		_, ok := ch.Game.Scripts[uint(id)]
		if !ok {
			ch.Send("A script with that ID could not be found.\r\n")
			return
		}

		/* For now, crudely do this from the host.  In the future we'll rewrite this shell command in JS anyway. */
		_, err = ch.Game.vm.RunString(fmt.Sprintf(`
			(function() {
				try {
					const ch = Golem.game.findPlayerByName("%s")[0];
					const script = Golem.game.scripts[%d];

					if(!ch || !ch.client || !script) {
						return;
					}

					Golem.StringEditor(ch.client,
						script.script,
						(_, string) => {
							ch.send("{WSaving script " + script.name + " (" + script.id + ")...{x\r\n");
							script.script = string;

							if(!script.save()) {
								ch.send("{RSave failed.{x\r\n");
							}
							
							ch.send("{WSaved, trying to re-evaluate script for exports...{x\r\n");

							try {
								var newExports = script.getExports();
								script.exports = newExports;

								ch.send("{Y" + JSON.stringify(newExports) + "\r\n");
								ch.send("{GScript exports updated in-place after a successful execution.{x\r\n");
							} catch(err) {
								ch.send("{RFailed to update script exports in place: " + err + "{x\r\n");
							}
						});
				} catch(err) {
				    Golem.game.broadcast(err);
				}
			})();
		`, ch.Name, id))
		if err != nil {
			ch.Send(fmt.Sprintf("Failed to edit this script: %v\r\n", err))
		}

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

		script, ok := ch.Game.Scripts[uint(id)]
		if !ok {
			ch.Send("A script with that ID could not be found.\r\n")
			return
		}

		err = ch.Game.DeleteScript(script)
		if err != nil {
			ch.Send(fmt.Sprintf("Something went wrong trying to delete this script: %v\r\n", err))
			return
		}

		ch.Send("Ok.\r\n")
		return

	default:
		ch.Send("Unrecognized command.\r\n")
	}
}

func do_goto(ch *Character, arguments string) {
	firstArgument, arguments := oneArgument(arguments)
	secondArgument, _ := oneArgument(arguments)

	if firstArgument == "plane" {
		id, err := strconv.Atoi(secondArgument)
		if err != nil || id <= 0 {
			ch.Send("Goto which plane ID?\r\n")
			return
		}

		var found *Plane = nil
		for iter := ch.Game.Planes.Head; iter != nil; iter = iter.Next {
			plane := iter.Value.(*Plane)

			if plane.Id == id {
				found = plane
				break
			}
		}

		if found == nil {
			ch.Send("No such plane.\r\n")
			return
		}

		destination := found.MaterializeRoom(0, 0, 0, true)

		if ch.Room != nil && ch.Room.Characters != nil {
			for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
				character := iter.Value.(*Character)
				if character != ch {
					character.Send(fmt.Sprintf("\r\n{W%s{W disappears in a puff of smoke.{x\r\n", ch.GetShortDescriptionUpper(character)))
				}
			}
		}

		if ch.Room != nil {
			ch.Room.removeCharacter(ch)
			destination.AddCharacter(ch)

			for iter := destination.Characters.Head; iter != nil; iter = iter.Next {
				character := iter.Value.(*Character)
				if character != ch {
					character.Send(fmt.Sprintf("\r\n{W%s{W appears in a puff of smoke.{x\r\n", ch.GetShortDescriptionUpper(character)))
				}
			}
		}

		do_look(ch, "")
		return
	}

	var room *Room = nil

	for iter := ch.Game.Characters.Head; iter != nil; iter = iter.Next {
		gch := iter.Value.(*Character)

		nameParts := strings.Split(gch.Name, " ")
		for _, part := range nameParts {
			if strings.Compare(strings.ToLower(part), firstArgument) == 0 {
				room = gch.Room
				break
			}
		}
	}

	if room == nil {
		id, err := strconv.Atoi(firstArgument)
		if err != nil || id <= 0 {
			ch.Send("Goto which room?\r\n")
			return
		}

		room, err = ch.Game.LoadRoomIndex(uint(id))
		if err != nil {
			ch.Send("No such room.\r\n")
			return
		}
	}

	if room == nil {
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
