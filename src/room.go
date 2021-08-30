/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

/* Fixed location IDs */
const RoomLimbo = 1
const RoomDeveloperLounge = 2

type Room struct {
	id uint

	name        string
	description string

	objects    *LinkedList
	resets     *LinkedList
	characters *LinkedList

	exit map[uint]*Exit
}

func (room *Room) addObject(obj *ObjectInstance) {
	room.objects.Insert(obj)
}

func (room *Room) removeObject(obj *ObjectInstance) {
	room.objects.Remove(obj)
}

func (room *Room) addCharacter(ch *Character) {
	room.characters.Insert(ch)
	ch.room = room
}

func (room *Room) removeCharacter(ch *Character) {
	room.characters.Remove(ch)
	ch.room = nil
}

func (room *Room) listObjectsToCharacter(ch *Character) {
	var output strings.Builder

	for iter := room.objects.head; iter != nil; iter = iter.next {
		obj := iter.value.(*ObjectInstance)

		output.WriteString(fmt.Sprintf("    {W%s{x\r\n", obj.longDescription))
	}

	ch.Send(output.String())
}

func (room *Room) listOtherRoomCharactersToCharacter(ch *Character) {
	var output strings.Builder

	for iter := room.characters.head; iter != nil; iter = iter.next {
		rch := iter.value.(*Character)

		if rch != ch {
			output.WriteString(fmt.Sprintf("{G%s{x\r\n", rch.getLongDescription(ch)))
		}
	}

	ch.Send(output.String())
}

func (game *Game) LoadRoomIndex(index uint) (*Room, error) {
	room, ok := game.world[index]
	if ok {
		return room, nil
	}

	row := game.db.QueryRow(`
		SELECT
			id,
			name,
			description
		FROM
			rooms
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`, index)

	room = &Room{}
	room.resets = NewLinkedList()
	room.objects = NewLinkedList()
	room.characters = NewLinkedList()
	room.exit = make(map[uint]*Exit)
	err := row.Scan(&room.id, &room.name, &room.description)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	game.world[room.id] = room
	return room, nil
}

func (game *Game) FixExits() error {
	log.Printf("Fixing exits.\r\n")

	rows, err := game.db.Query(`
		SELECT
			id,
			room_id,
			to_room_id,
			direction,
			flags
		FROM
			exits
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var err error

		exit := &Exit{}

		var roomId int
		var toRoomId int

		rows.Scan(&exit.id, &roomId, &toRoomId, &exit.direction, &exit.flags)
		exit.to, err = game.LoadRoomIndex(uint(toRoomId))
		if err != nil {
			continue
		}

		from, err := game.LoadRoomIndex(uint(roomId))
		if err != nil {
			continue
		}

		from.exit[exit.direction] = exit
	}

	return nil
}
