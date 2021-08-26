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

	exit map[uint]*Exit
}

var World map[uint]*Room

func LoadRoomIndex(index uint) *Room {
	room, ok := World[index]
	if ok {
		return room
	}

	return nil
}
