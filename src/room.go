/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

type Room struct {
	id uint

	name        string
	description string

	characters map[*Character]bool
	exit       map[uint]*Exit
}

func (game *Game) LoadRoomIndex(index uint) *Room {
	room, ok := game.world[index]
	if ok {
		return room
	}

	return nil
}
