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
	"time"
)

type AwayFromKeyboard struct {
	startedAt time.Time
	message   string
}

func do_afk(ch *Character, arguments string) {
	var reason string = "currently away"

	if len(arguments) != 0 {
		reason = string(arguments)
	}

	if ch.afk != nil {
		ch.Send("{GYou have returned from AFK.{x\r\n")
		ch.afk = nil
		return
	}

	ch.afk = &AwayFromKeyboard{
		startedAt: time.Now(),
		message:   string(reason),
	}

	ch.Send(fmt.Sprintf("{GYou are now AFK: %s{x\r\n", ch.afk.message))
}

/* say will be room-specific */
func do_say(ch *Character, arguments string) {
	var buf strings.Builder

	if len(arguments) == 0 {
		ch.Send("{CSay what?{x\r\n")
		return
	}

	ch.Send(fmt.Sprintf("{CYou say \"%s{C\"{x\r\n", arguments))

	buf.WriteString(fmt.Sprintf("\r\n{C%s says \"%s{C\"{x\r\n", ch.Name, arguments))
	output := buf.String()

	if ch.Room != nil {
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch != ch {
				rch.Send(output)
			}
		}
	}
}

func do_ooc(ch *Character, arguments string) {
	var buf strings.Builder

	if len(arguments) == 0 {
		ch.Send("{MWhat message do you wish to globally send out-of-character?{x\r\n")
		return
	}

	buf.WriteString(fmt.Sprintf("\r\n{M[OOC] %s: %s{x\r\n", ch.Name, arguments))
	output := buf.String()

	for client := range ch.game.clients {
		if client.character != nil && client.connectionState == ConnectionStatePlaying {
			client.character.Send(output)
		}
	}
}

func do_save(ch *Character, arguments string) {
	result := ch.Save()
	if !result {
		ch.Send("A strange force prevents you from saving.\r\n")
		return
	}

	ch.Send("Saved.\r\n")
}

func do_quit(ch *Character, arguments string) {
	ch.Save()

	/* If this character is leading a group, disband it */
	if ch.Group != nil {
		if ch.Leader == ch {
			ch.DisbandGroup()
		} else {
			ch.Leader.Group.Remove(ch)
		}
	}

	if ch.Room != nil {
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			character := iter.Value.(*Character)

			if character != ch {
				character.Send(fmt.Sprintf("{W%s{W has quit the game.{x\r\n", ch.GetShortDescriptionUpper(character)))
			}
		}

		ch.Room.removeCharacter(ch)
	}

	ch.game.Characters.Remove(ch)
	ch.client.connectionState = ConnectionStateNone
	ch.Send("{WLeaving for the real world...{x\r\n")

	go func() {
		/* Allow output to flush */
		<-time.After(500 * time.Millisecond)
		ch.client.close <- true
		ch.client.conn.Close()
	}()
}

func (ch *Character) DisbandGroup() {
	ch.Send("{WYou disband your group.{x\r\n")

	for iter := ch.Group.Head; iter != nil; iter = iter.Next {
		gch := iter.Value.(*Character)
		gch.Group = nil
		gch.Leader = nil

		if gch != ch {
			gch.Send(fmt.Sprintf("{W%s{W disbanded the group.{x\r\n", ch.GetShortDescriptionUpper(gch)))
		}
	}
}

func do_group(ch *Character, arguments string) {
	if len(arguments) < 1 {
		if ch.Group == nil {
			ch.Send("You aren't currently in a group.\r\n")
			return
		}

		var output strings.Builder

		output.WriteString(fmt.Sprintf("{W%s{W's group:{x\r\n", ch.Leader.GetShortDescriptionUpper(ch)))

		for iter := ch.Group.Head; iter != nil; iter = iter.Next {
			gch := iter.Value.(*Character)

			output.WriteString(fmt.Sprintf("[%2d %-8s] %-14s %5d/%5dhp %5d/%5dm %5d/%5dst\r\n",
				gch.level,
				gch.job.DisplayName,
				gch.GetShortDescriptionUpper(ch),
				gch.health,
				gch.maxHealth,
				gch.mana,
				gch.maxMana,
				gch.stamina,
				gch.maxStamina))
		}

		ch.Send(output.String())
		return
	}

	arg, _ := oneArgument(arguments)
	target := ch.FindCharacterInRoom(arg)

	if target == nil {
		ch.Send("They aren't here.\r\n")
		return
	} else if target == ch && ch.Leader == ch {
		ch.DisbandGroup()
		return
	} else if target == ch && ch.Leader != ch {
		ch.Send(fmt.Sprintf("{WYou leave %s{W's group.{x\r\n", ch.Leader.GetShortDescription(ch)))

		for iter := ch.Group.Head; iter != nil; iter = iter.Next {
			gch := iter.Value.(*Character)

			if gch != ch {
				gch.Send(fmt.Sprintf("{W%s{W leaves the group.{x\r\n", ch.GetShortDescriptionUpper(gch)))
			}
		}

		ch.Leader.Group.Remove(ch)
		ch.Group = nil
		return
	}

	if target.Following != ch {
		ch.Send("They aren't following you.\r\n")
		return
	}

	if ch.Group == nil {
		ch.Group = NewLinkedList()
		ch.Group.Insert(ch)
		ch.Leader = ch
	}

	if ch.Group.Contains(target) {
		ch.Send(fmt.Sprintf("{WYou remove %s{W from your group.{x\r\n", target.GetShortDescription(ch)))
		target.Send(fmt.Sprintf("{W%s{W removes you from their group.{x\r\n", ch.GetShortDescriptionUpper(target)))

		ch.Group.Remove(target)
		target.Leader = nil
	} else {
		ch.Send(fmt.Sprintf("{W%s{W joins your group.{x\r\n", target.GetShortDescriptionUpper(ch)))
		target.Send(fmt.Sprintf("{WYou join %s{W's group.{x\r\n", ch.GetShortDescription(target)))

		ch.Group.Insert(target)

		target.Leader = ch
		target.Group = ch.Group
	}
}
