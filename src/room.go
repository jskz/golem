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

type Room struct {
	id uint

	name        string
	description string

	resets     *LinkedList
	characters map[*Character]bool
	exit       map[uint]*Exit
}

func (room *Room) addCharacter(ch *Character) {
	_, ok := room.characters[ch]
	if ok {
		return
	}

	room.characters[ch] = true
	ch.room = room
}

func (room *Room) removeCharacter(ch *Character) {
	delete(room.characters, ch)
	ch.room = nil
}

func (room *Room) listObjectsToCharacter(ch *Character) {

}

func (room *Room) listOtherRoomCharactersToCharacter(ch *Character) {
	var output strings.Builder

	for rch := range room.characters {
		if rch != ch {
			output.WriteString(fmt.Sprintf("{G%s{x\r\n", rch.getLongDescription()))
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
	room.characters = make(map[*Character]bool)
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
