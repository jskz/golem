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

	buf.WriteString(fmt.Sprintf("\r\n{C%s says \"%s{C\"{x\r\n", ch.name, arguments))
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

	buf.WriteString(fmt.Sprintf("\r\n{M[OOC] %s: %s{x\r\n", ch.name, arguments))
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
	ch.client.conn.Close()
}
