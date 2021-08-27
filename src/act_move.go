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
)

var ReverseDirection = map[int]int{
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
	return nil
}

func (ch *Character) move(direction uint) bool {
	return false
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