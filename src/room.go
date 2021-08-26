/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "database/sql"

type Room struct {
	id uint

	name        string
	description string

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
	room.exit = make(map[uint]*Exit)
	err := row.Scan(&room.id, &room.name, &room.description)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return room, nil
}
