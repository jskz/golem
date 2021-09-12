/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "fmt"

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

var ReverseDirection = map[uint]uint{
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
	const MovementCost = 2 /* update with terrain-based cost */

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

	if ch.stamina-MovementCost < 0 {
		ch.Send("{DYou are too exhausted to move!{x\r\n")
		return false
	}

	ch.stamina -= MovementCost

	/* Is the exit closed, etc. */
	from := ch.room
	from.removeCharacter(ch)
	for iter := from.characters.head; iter != nil; iter = iter.next {
		character := iter.value.(*Character)
		character.Send(fmt.Sprintf("{W%s{W leaves %s.{x\r\n", ch.getShortDescriptionUpper(character), ExitName[direction]))
	}

	exit.to.addCharacter(ch)
	for iter := exit.to.characters.head; iter != nil; iter = iter.next {
		character := iter.value.(*Character)
		if character != ch {
			character.Send(fmt.Sprintf("{W%s{W arrives from %s.{x\r\n", ch.getShortDescriptionUpper(character), ExitName[ReverseDirection[direction]]))
		}
	}

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
