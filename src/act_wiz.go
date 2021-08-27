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
)

func do_exec(ch *Character, arguments string) {
	if ch.client != nil {
		value, err := ch.client.game.vm.RunString(arguments)

		if err != nil {
			ch.Send(fmt.Sprintf("{R\r\nError: %s{x.\r\n", err.Error()))
			return
		}

		ch.Send(fmt.Sprintf("{w\r\n%s{x\r\n", value.String()))
	}
}

func do_shutdown(ch *Character, arguments string) {
	if ch.client != nil {
		ch.client.game.shutdownRequest <- true
	}
}

func do_goto(ch *Character, arguments string) {
	id, err := strconv.Atoi(arguments)
	if err != nil || id <= 0 {
		ch.Send("Goto which room ID?\r\n")
		return
	}

	room, err := ch.client.game.LoadRoomIndex(uint(id))
	if err != nil || room == nil {
		ch.Send("No such room.\r\n")
		return
	}

	if ch.room != nil {
		ch.room.removeCharacter(ch)
	}

	room.addCharacter(ch)
	do_look(ch, "")
}
