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

func (ch *Character) move(direction uint, follow bool) bool {
	const MovementCost = 2 /* update with terrain-based cost */

	if ch.isFighting() {
		ch.Send("{RYou are in the middle of fighting!{x\r\n")
		return false
	}

	if ch.casting != nil {
		ch.Send("{RYou are focused on casting a magical spell and cannot move!{x\r\n")
		return false
	}

	if ch.Room == nil {
		ch.Send("{RAlas, you cannot go that way.{x\r\n")
		return false
	}

	exit := ch.Room.getExit(direction)
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
	from := ch.Room
	from.removeCharacter(ch)
	for iter := from.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		character.Send(fmt.Sprintf("{W%s{W leaves %s.{x\r\n", ch.GetShortDescriptionUpper(character), ExitName[direction]))
	}

	exit.to.addCharacter(ch)
	for iter := exit.to.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		if character != ch {
			character.Send(fmt.Sprintf("{W%s{W arrives from %s.{x\r\n", ch.GetShortDescriptionUpper(character), ExitName[ReverseDirection[direction]]))
		}
	}

	do_look(ch, "")

	if exit.to == from {
		return true
	}

	for iter := from.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)

		if character.following == ch {
			character.Send(fmt.Sprintf("{WYou follow %s{W.{x\r\n", ch.GetShortDescription(character)))
			character.move(direction, true)
		}
	}

	return true
}

func do_follow(ch *Character, arguments string) {
	if ch.Room == nil {
		return
	}

	if len(arguments) < 1 {
		ch.Send("Follow who?\r\n")
		return
	}

	var target *Character = ch.FindCharacterInRoom(arguments)

	if target == nil {
		ch.Send("No such target.  Follow who?\r\n")
		return
	}

	if target == ch && ch.following == nil {
		ch.Send("You are already following yourself.\r\n")
		return
	}

	if target == ch && ch.following != nil {
		ch.Send(fmt.Sprintf("You stop following %s{x.\r\n", ch.following.GetShortDescription(ch)))
		ch.following = nil
		return
	}

	if ch.following != nil {
		ch.Send("You are already following somebody else.  Follow yourself first.\r\n")
		return
	}

	ch.Send(fmt.Sprintf("You start following %s{x.\r\n", target.GetShortDescription(ch)))
	ch.following = target
}

func do_north(ch *Character, arguments string) {
	ch.move(DirectionNorth, false)
}

func do_east(ch *Character, arguments string) {
	ch.move(DirectionEast, false)
}

func do_south(ch *Character, arguments string) {
	ch.move(DirectionSouth, false)
}

func do_west(ch *Character, arguments string) {
	ch.move(DirectionWest, false)
}

func do_up(ch *Character, arguments string) {
	ch.move(DirectionUp, false)
}

func do_down(ch *Character, arguments string) {
	ch.move(DirectionDown, false)
}
