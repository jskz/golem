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

	if ch.room != nil {
		for iter := ch.room.characters.head; iter != nil; iter = iter.next {
			rch := iter.value.(*Character)

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

	for client := range ch.client.game.clients {
		if client.character != nil && client.connectionState == ConnectionStatePlaying {
			client.character.Send(output)
		}
	}
}

func do_save(ch *Character, arguments string) {
	ch.Save()

	ch.Send("Saved.\r\n")
}

func do_quit(ch *Character, arguments string) {
	var output strings.Builder

	sessionDuration := time.Since(ch.client.sessionStartedAt)
	seconds := int(sessionDuration.Seconds())

	output.WriteString("\r\n{RYou are leaving the game world.\r\n")
	output.WriteString(fmt.Sprintf("You were connected for %d seconds.{x\r\n", seconds))

	ch.Save()

	if ch.room != nil {
		for iter := ch.room.characters.head; iter != nil; iter = iter.next {
			character := iter.value.(*Character)

			if character != ch {
				character.Send(fmt.Sprintf("{W%s{W has quit the game.{x\r\n", ch.getShortDescriptionUpper(character)))
			}
		}

		ch.room.removeCharacter(ch)
	}

	ch.Send(output.String())
	ch.client.conn.Close()
}
