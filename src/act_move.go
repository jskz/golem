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
	DirectionNorth:     "north",
	DirectionEast:      "east",
	DirectionSouth:     "south",
	DirectionWest:      "west",
	DirectionUp:        "up",
	DirectionDown:      "down",
	DirectionNortheast: "northeast",
	DirectionSoutheast: "southeast",
	DirectionSouthwest: "southwest",
	DirectionNorthwest: "northwest",
}

var ExitCompassName = map[uint]string{
	DirectionNorth:     "N",
	DirectionEast:      "E",
	DirectionSouth:     "S",
	DirectionWest:      "W",
	DirectionUp:        "U",
	DirectionDown:      "D",
	DirectionNortheast: "NE",
	DirectionSoutheast: "SE",
	DirectionSouthwest: "SW",
	DirectionNorthwest: "NW",
}

var ReverseDirection = map[uint]uint{
	DirectionNorth:     DirectionSouth,
	DirectionEast:      DirectionWest,
	DirectionSouth:     DirectionNorth,
	DirectionWest:      DirectionEast,
	DirectionUp:        DirectionDown,
	DirectionDown:      DirectionUp,
	DirectionNortheast: DirectionSouthwest,
	DirectionSoutheast: DirectionNorthwest,
	DirectionSouthwest: DirectionNortheast,
	DirectionNorthwest: DirectionSoutheast,
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
	TerrainTypeCaveWall              = 1
	TerrainTypeCaveDeepWall1         = 2
	TerrainTypeCaveDeepWall2         = 3
	TerrainTypeCaveDeepWall3         = 4
	TerrainTypeCaveDeepWall4         = 5
	TerrainTypeCaveDeepWall5         = 6
	TerrainTypeCaveTunnel            = 7
	TerrainTypeOcean                 = 8
	TerrainTypeOverworldCityExterior = 9
	TerrainTypeOverworldCityInterior = 10
	TerrainTypeOverworldCityEntrance = 11
	TerrainTypePlains                = 12
	TerrainTypeField                 = 13
	TerrainTypeShore                 = 14
	TerrainTypeShallowWater          = 15
	TerrainTypeLightForest           = 16
	TerrainTypeDenseForest           = 17
	TerrainTypeHills                 = 18
	TerrainTypeMountains             = 19
	TerrainTypeSnowcappedMountains   = 20
)

type Terrain struct {
	Id           int    `json:"id"`
	Name         string `json:"name"`
	GlyphColour  string `json:"glyphColour"`
	MapGlyph     string `json:"mapGlyph"`
	MovementCost int    `json:"movementCost"`
	Flags        int    `json:"flags"`
}

type Exit struct {
	Id        uint  `json:"id"`
	Direction uint  `json:"direction"`
	To        *Room `json:"to"`
	Flags     int   `json:"flags"`
}

func (game *Game) NewExit(direction uint, to *Room, flags int) *Exit {
	return &Exit{
		Id:        0,
		To:        to,
		Direction: direction,
		Flags:     flags,
	}
}

func (room *Room) getExit(direction uint) *Exit {
	return room.Exit[direction]
}

func (ch *Character) move(direction uint, follow bool) bool {
	const MovementCost = 2 /* update with terrain-based cost */

	if ch.isFighting() {
		ch.Send("{RYou are in the middle of fighting!{x\r\n")
		return false
	}

	if ch.Casting != nil {
		ch.Send("{RYou are focused on casting a magical spell and cannot move!{x\r\n")
		return false
	}

	if ch.Room == nil {
		ch.Send("{RAlas, you cannot go that way.{x\r\n")
		return false
	}

	exit := ch.Room.getExit(direction)
	if exit == nil || exit.To == nil {
		ch.Send("{RAlas, you cannot go that way.{x\r\n")
		return false
	}

	if exit.Flags&EXIT_CLOSED != 0 {
		ch.Send("{RIt is closed.{x\r\n")
		return false
	}

	if ch.Stamina-MovementCost < 0 {
		ch.Send("{DYou are too exhausted to move!{x\r\n")
		return false
	}

	ch.Stamina -= MovementCost

	// Is the exit closed, etc.
	from := ch.Room
	from.removeCharacter(ch)
	for iter := from.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		character.Send(fmt.Sprintf("{W%s{W leaves %s.{x\r\n", ch.GetShortDescriptionUpper(character), ExitName[direction]))
	}

	if from.script != nil {
		from.script.tryEvaluate("onRoomLeave", ch.Game.vm.ToValue(from), ch.Game.vm.ToValue(ch))
	}

	// If destination room is planar, then try to fully materialize on the room that the player is about to move into
	if exit.To.Flags&ROOM_PLANAR != 0 {
		exit.To = exit.To.Plane.MaterializeRoom(exit.To.X, exit.To.Y, exit.To.Z, true)
	}

	exit.To.AddCharacter(ch)
	for iter := exit.To.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		if character != ch {
			character.Send(fmt.Sprintf("{W%s{W arrives from %s.{x\r\n", ch.GetShortDescriptionUpper(character), ExitName[ReverseDirection[direction]]))
		}
	}

	do_look(ch, "")

	if exit.To == from {
		return true
	}

	for iter := from.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)

		if character.Following == ch {
			character.Send(fmt.Sprintf("{WYou follow %s{W.{x\r\n", ch.GetShortDescription(character)))
			character.move(direction, true)
		}
	}

	if exit.To.script != nil {
		exit.To.script.tryEvaluate("onRoomEnter", ch.Game.vm.ToValue(exit.To), ch.Game.vm.ToValue(ch))
	}

	/* Aggro check... */
	for iter := exit.To.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)

		if character != ch {
			/* If the entering player is a PC, this is a hostile NPC, and that hostile NPC is not currently preoccupied with another combat, then let's rum	ble. */
			if (ch.Flags&CHAR_IS_PLAYER != 0) && (character.Flags&CHAR_IS_PLAYER == 0) && (character.Flags&CHAR_AGGRESSIVE != 0) && (character.Fighting == nil) {
				do_kill(character, ch.Name)
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

	if exit == nil || exit.Flags&EXIT_IS_DOOR == 0 {
		ch.Send("You can't close that.\r\n")
		return
	}

	if exit.Flags&EXIT_CLOSED != 0 {
		ch.Send("It's already closed.\r\n")
		return
	}

	exit.Flags |= EXIT_CLOSED
	exit.To.Exit[ReverseDirection[exit.Direction]].Flags |= EXIT_CLOSED

	ch.Send("You close the door.\r\n")

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			rch.Send(fmt.Sprintf("{W%s{W closes the door %s.{x\r\n", ch.GetShortDescriptionUpper(rch), ExitName[exit.Direction]))
		}
	}

	for iter := exit.To.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		character.Send(fmt.Sprintf("{WThe %s door closes.{x\r\n", ExitName[ReverseDirection[exit.Direction]]))
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

	if exit == nil || exit.Flags&EXIT_IS_DOOR == 0 {
		obj := ch.findObjectOnSelf(args)

		if obj == nil {
			obj = ch.findObjectInRoom(args)
		}

		if obj == nil || obj.Flags&ITEM_CLOSEABLE == 0 {
			ch.Send("You can't open that.\r\n")
			return
		} else {
			if obj.Flags&ITEM_CLOSED == 0 {
				ch.Send("It's already open.\r\n")
				return
			}

			obj.Flags &= ^ITEM_CLOSED
			ch.Send(fmt.Sprintf("You open %s{x.\r\n", obj.GetShortDescription(ch)))

			for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
				rch := iter.Value.(*Character)

				if !rch.IsEqual(ch) {
					rch.Send(fmt.Sprintf("{W%s{W opens %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
				}
			}

			return
		}
	}

	if exit.Flags&EXIT_CLOSED == 0 {
		ch.Send("It isn't closed.\r\n")
		return
	} else if exit.Flags&EXIT_LOCKED != 0 {
		ch.Send("It's locked.\r\n")
		return
	}

	if exit.To == nil || exit.To.Exit[ReverseDirection[exit.Direction]] == nil {
		ch.Send("{DA mysterious force prevents you from opening the door.{x\r\n")
		return
	}

	exit.Flags &= ^EXIT_CLOSED
	exit.To.Exit[ReverseDirection[exit.Direction]].Flags &= ^EXIT_CLOSED

	ch.Send("You open the door.\r\n")

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if !rch.IsEqual(ch) {
			rch.Send(fmt.Sprintf("{W%s{W opens the door %s.{x\r\n", ch.GetShortDescriptionUpper(rch), ExitName[exit.Direction]))
		}
	}

	for iter := exit.To.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		character.Send(fmt.Sprintf("{WThe %s door opens.{x\r\n", ExitName[ReverseDirection[exit.Direction]]))
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
