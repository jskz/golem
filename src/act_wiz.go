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

func do_exec(ch *Character, arguments string) {
	if ch.client != nil {
		value, err := ch.game.vm.RunString(arguments)

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

	for iter := ch.game.zones.Head; iter != nil; iter = iter.Next {
		zone := iter.Value.(*Zone)

		minutesSinceZoneReset := int(time.Since(zone.lastReset).Minutes())

		output.WriteString(fmt.Sprintf("%-4d %-35s [%-5d-%-5d] %d/%d\r\n",
			zone.id,
			zone.name,
			zone.low,
			zone.high,
			zone.resetFrequency,
			minutesSinceZoneReset))
	}

	output.WriteString("{x")
	ch.Send(output.String())
}

func do_mem(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString("{YUsage statistics:\r\n")
	output.WriteString(fmt.Sprintf("%-15s %-6d\r\n", "Characters", ch.game.characters.Count))
	output.WriteString(fmt.Sprintf("%-15s %-6d\r\n", "Jobs", Jobs.Count))
	output.WriteString(fmt.Sprintf("%-15s %-6d\r\n", "Races", Races.Count))
	output.WriteString(fmt.Sprintf("%-15s %-6d{x\r\n", "Zones", ch.game.zones.Count))

	ch.Send(output.String())
}

func do_mlist(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Displaying all %d character instances in the world:\r\n", ch.game.characters.Count))

	for iter := ch.game.characters.Head; iter != nil; iter = iter.Next {
		wch := iter.Value.(*Character)

		if wch.flags&CHAR_IS_PLAYER != 0 {
			if wch.client != nil {
				output.WriteString(fmt.Sprintf("{G%s@%s{x\r\n", wch.name, wch.client.conn.RemoteAddr().String()))
			} else {
				output.WriteString(fmt.Sprintf("{G%s@DISCONNECTED{x\r\n", wch.name))
			}
		} else {
			output.WriteString(fmt.Sprintf("%s{x (id#%d)\r\n", wch.getShortDescriptionUpper(ch), wch.id))
		}
	}

	ch.Send(output.String())
}

func do_purge(ch *Character, arguments string) {
	if ch.Room == nil {
		return
	}

	for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)
		if rch == ch || rch.client != nil || rch.flags&CHAR_IS_PLAYER != 0 {
			continue
		}

		ch.Room.characters.Remove(rch)
	}

	for {
		if ch.Room.objects.Head == nil {
			break
		}

		ch.Room.objects.Remove(ch.Room.objects.Head.Value)
	}

	ch.Send("You have purged the contents of the room.\r\n")
}

func do_peace(ch *Character, arguments string) {
	if ch.Room == nil || ch.client == nil {
		return
	}

	for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		rch.flags &= ^CHAR_AGGRESSIVE
		rch.Fighting = nil
		rch.Combat = nil
	}

	ch.Send("Ok.\r\n")
}

func do_shutdown(ch *Character, arguments string) {
	if ch.client != nil {
		ch.game.shutdownRequest <- true
	}
}

func do_goto(ch *Character, arguments string) {
	id, err := strconv.Atoi(arguments)
	if err != nil || id <= 0 {
		ch.Send("Goto which room ID?\r\n")
		return
	}

	room, err := ch.game.LoadRoomIndex(uint(id))
	if err != nil || room == nil {
		ch.Send("No such room.\r\n")
		return
	}

	if ch.Room != nil {
		for iter := room.characters.Head; iter != nil; iter = iter.Next {
			character := iter.Value.(*Character)
			if character != ch {
				character.Send(fmt.Sprintf("\r\n{W%s{W disappears in a puff of smoke.{x\r\n", ch.getShortDescriptionUpper(character)))
			}
		}

		ch.Room.removeCharacter(ch)
	}

	room.addCharacter(ch)

	for iter := room.characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		if character != ch {
			character.Send(fmt.Sprintf("\r\n{W%s{W appears in a puff of smoke.{x\r\n", ch.getShortDescriptionUpper(character)))
		}
	}

	do_look(ch, "")
}
