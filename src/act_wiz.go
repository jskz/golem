/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "strconv"

func do_shutdown(ch *Character, arguments string) {
	if ch.client != nil {
		ch.client.game.shutdownRequest <- true
	}
}

func do_goto(ch *Character, arguments string) {
	id, err := strconv.Atoi(arguments)
	if err != nil || id <= 0 {
		ch.send("Goto which room ID?\r\n")
		return
	}

	room, err := ch.client.game.LoadRoomIndex(uint(id))
	if err != nil || room == nil {
		ch.send("No such room.\r\n")
		return
	}

	if ch.room != nil {
		ch.room.removeCharacter(ch)
		room.addCharacter(ch)
	}
}
