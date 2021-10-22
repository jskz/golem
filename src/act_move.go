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
)

const (
	DirectionNorth     = 0
	DirectionEast      = 1
	DirectionSouth     = 2
	DirectionWest      = 3
	DirectionUp        = 4
	DirectionDown      = 5
	DirectionNortheast = 6
	DirectionSoutheast = 7
	DirectionSouthwest = 8
	DirectionNorthwest = 9
	DirectionMax       = 10
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

const (
	EXIT_IS_DOOR = 1
	EXIT_CLOSED  = 1 << 1
	EXIT_LOCKED  = 1 << 2
	EXIT_HIDDEN  = 1 << 3
)

const (
	TERRAIN_IMPASSABLE    = 1
	TERRAIN_SHALLOW_WATER = 1 << 1
	TERRAIN_DEEP_WATER    = 1 << 2
)

const (
	TerrainTypeCaveWall      = 1
	TerrainTypeCaveDeepWall1 = 2
	TerrainTypeCaveDeepWall2 = 3
	TerrainTypeCaveDeepWall3 = 4
	TerrainTypeCaveDeepWall4 = 5
	TerrainTypeCaveDeepWall5 = 6
	TerrainTypeCaveTunnel    = 7
)

type Terrain struct {
	id           int
	name         string
	mapGlyph     string
	movementCost int
	flags        int
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

	if exit.flags&EXIT_CLOSED != 0 {
		ch.Send("{RIt is closed.{x\r\n")
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

	if from.script != nil {
		from.script.tryEvaluate("onRoomLeave", ch.game.vm.ToValue(from), ch.game.vm.ToValue(ch))
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

		if character.Following == ch {
			character.Send(fmt.Sprintf("{WYou follow %s{W.{x\r\n", ch.GetShortDescription(character)))
			character.move(direction, true)
		}
	}

	if exit.to.script != nil {
		exit.to.script.tryEvaluate("onRoomEnter", ch.game.vm.ToValue(exit.to), ch.game.vm.ToValue(ch))
	}

	/* Aggro check... */
	for iter := exit.to.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)

		if character != ch {
			/* If the entering player is a PC, this is a hostile NPC, and that hostile NPC is not currently preoccupied with another combat, then let's rum	ble. */
			if (ch.flags&CHAR_IS_PLAYER != 0) && (character.flags&CHAR_IS_PLAYER == 0) && (character.flags&CHAR_AGGRESSIVE != 0) {
				do_kill(character, ch.name)
			}
		}
	}

	return true
}

func do_close(ch *Character, arguments string) {
	/* Will be able to close containers and closeable exits in the room - for now just exits */
	if ch.Room == nil {
		return
	}

	if len(arguments) < 1 {
		ch.Send("Close what?\r\n")
		return
	}

	args := strings.ToLower(arguments)
	var exit *Exit = nil

	if args == "n" || args == "north" {
		exit = ch.Room.getExit(DirectionNorth)
	} else if args == "e" || args == "east" {
		exit = ch.Room.getExit(DirectionEast)
	} else if args == "s" || args == "south" {
		exit = ch.Room.getExit(DirectionSouth)
	} else if args == "w" || args == "west" {
		exit = ch.Room.getExit(DirectionWest)
	} else if args == "u" || args == "up" {
		exit = ch.Room.getExit(DirectionUp)
	} else if args == "d" || args == "down" {
		exit = ch.Room.getExit(DirectionDown)
	} else {
		ch.Send("Close what?\r\n")
		return
	}

	if exit == nil || exit.flags&EXIT_IS_DOOR == 0 {
		ch.Send("You can't close that.\r\n")
		return
	}

	if exit.flags&EXIT_CLOSED != 0 {
		ch.Send("It's already closed.\r\n")
		return
	}

	exit.flags |= EXIT_CLOSED
	exit.to.exit[ReverseDirection[exit.direction]].flags |= EXIT_CLOSED

	ch.Send("You close the door.\r\n")

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			rch.Send(fmt.Sprintf("{W%s{W closes the door %s.{x\r\n", ch.GetShortDescriptionUpper(rch), ExitName[exit.direction]))
		}
	}

	for iter := exit.to.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		character.Send(fmt.Sprintf("{WThe %s door closes.{x\r\n", ExitName[ReverseDirection[exit.direction]]))
	}
}

func do_open(ch *Character, arguments string) {
	/* Will be able to open containers and closeable exits in the room - for now just exits */
	if ch.Room == nil {
		return
	}

	if len(arguments) < 1 {
		ch.Send("Open what?\r\n")
		return
	}

	args := strings.ToLower(arguments)
	var exit *Exit = nil

	if args == "n" || args == "north" {
		exit = ch.Room.getExit(DirectionNorth)
	} else if args == "e" || args == "east" {
		exit = ch.Room.getExit(DirectionEast)
	} else if args == "s" || args == "south" {
		exit = ch.Room.getExit(DirectionSouth)
	} else if args == "w" || args == "west" {
		exit = ch.Room.getExit(DirectionWest)
	} else if args == "u" || args == "up" {
		exit = ch.Room.getExit(DirectionUp)
	} else if args == "d" || args == "down" {
		exit = ch.Room.getExit(DirectionDown)
	} else {
		ch.Send("Open what?\r\n")
		return
	}

	if exit == nil || exit.flags&EXIT_IS_DOOR == 0 {
		ch.Send("You can't open that.\r\n")
		return
	}

	if exit.flags&EXIT_CLOSED == 0 {
		ch.Send("It isn't closed.\r\n")
		return
	} else if exit.flags&EXIT_LOCKED != 0 {
		ch.Send("It's locked.\r\n")
		return
	}

	exit.flags &= ^EXIT_CLOSED
	exit.to.exit[ReverseDirection[exit.direction]].flags &= ^EXIT_CLOSED

	ch.Send("You open the door.\r\n")

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			rch.Send(fmt.Sprintf("{W%s{W opens the door %s.{x\r\n", ch.GetShortDescriptionUpper(rch), ExitName[exit.direction]))
		}
	}

	for iter := exit.to.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		character.Send(fmt.Sprintf("{WThe %s door opens.{x\r\n", ExitName[ReverseDirection[exit.direction]]))
	}
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

	if target == ch && ch.Following == nil {
		ch.Send("You are already following yourself.\r\n")
		return
	}

	if target == ch && ch.Following != nil {
		ch.Send(fmt.Sprintf("{WYou stop following %s{W.{x\r\n", ch.Following.GetShortDescription(ch)))
		ch.Following.Send(fmt.Sprintf("{W%s{W stops following you.{x", ch.GetShortDescriptionUpper(ch.Following)))
		ch.Following = nil
		return
	}

	if ch.Following != nil {
		ch.Send("You are already following somebody else.  Follow yourself first.\r\n")
		return
	}

	ch.Send(fmt.Sprintf("{WYou start following %s{x.\r\n", target.GetShortDescription(ch)))
	target.Send(fmt.Sprintf("{W%s{W starts following you.{x\r\n", ch.GetShortDescriptionUpper(target)))
	ch.Following = target
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
