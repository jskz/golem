/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

const (
	DirectionNorth = 0
	DirectionEast  = 1
	DirectionSouth = 2
	DirectionWest  = 3
	DirectionUp    = 4
	DirectionDown  = 5
	DirectionMax   = 6
)

var ExitName = map[uint]string{
	DirectionNorth: "north",
	DirectionEast:  "east",
	DirectionSouth: "south",
	DirectionWest:  "west",
	DirectionUp:    "up",
	DirectionDown:  "down",
}

var ExitCompassName = map[uint]string{
	DirectionNorth: "N",
	DirectionEast:  "E",
	DirectionSouth: "S",
	DirectionWest:  "W",
	DirectionUp:    "U",
	DirectionDown:  "D",
}

var ReverseDirection = map[uint]int{
	DirectionNorth: DirectionSouth,
	DirectionEast:  DirectionWest,
	DirectionSouth: DirectionNorth,
	DirectionWest:  DirectionEast,
	DirectionUp:    DirectionDown,
	DirectionDown:  DirectionUp,
}

type Exit struct {
	id        uint
	direction uint
	to        *Room
	flags     int
}

func (room *Room) getExit(direction uint) *Exit {
	return room.exit[direction]
}

func (ch *Character) move(direction uint) bool {
	if ch.isFighting() {
		ch.Send("{RYou are in the middle of fighting!{x\r\n")
		return false
	}

	if ch.room == nil {
		ch.Send("{RAlas, you cannot go that way.{x\r\n")
		return false
	}

	exit := ch.room.getExit(direction)
	if exit == nil || exit.to == nil {
		ch.Send("{RAlas, you cannot go that way.{x\r\n")
		return false
	}

	/* Is the exit closed, etc. */

	ch.room.removeCharacter(ch)
	exit.to.addCharacter(ch)
	do_look(ch, "")
	return true
}

func do_north(ch *Character, arguments string) {
	ch.move(DirectionNorth)
}

func do_east(ch *Character, arguments string) {
	ch.move(DirectionEast)
}

func do_south(ch *Character, arguments string) {
	ch.move(DirectionSouth)
}

func do_west(ch *Character, arguments string) {
	ch.move(DirectionWest)
}

func do_up(ch *Character, arguments string) {
	ch.move(DirectionUp)
}

func do_down(ch *Character, arguments string) {
	ch.move(DirectionDown)
}
