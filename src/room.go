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
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/dop251/goja"
)

/* Fixed location IDs */
const RoomLimbo = 1
const RoomDeveloperLounge = 2

/* Room flag types */
const (
	ROOM_PERSISTENT = 1
	ROOM_VIRTUAL    = 1 << 1
	ROOM_SAFE       = 1 << 2
)

type Room struct {
	game *Game

	Id   uint `json:"id"`
	zone *Zone

	flags   int
	virtual bool
	cell    *MazeCell

	name        string
	description string

	objects    *LinkedList
	resets     *LinkedList
	Characters *LinkedList `json:"characters"`

	exit map[uint]*Exit
}

func (room *Room) addObject(obj *ObjectInstance) {
	room.objects.Insert(obj)

	obj.inside = nil
	obj.carriedBy = nil
	obj.inRoom = room
}

func (room *Room) removeObject(obj *ObjectInstance) {
	room.objects.Remove(obj)

	obj.inRoom = nil
}

func (room *Room) addCharacter(ch *Character) {
	room.Characters.Insert(ch)

	ch.Room = room
}

func (room *Room) removeCharacter(ch *Character) {
	room.Characters.Remove(ch)
	ch.Room = nil
}

func (room *Room) listObjectsToCharacter(ch *Character) {
	var output strings.Builder

	for iter := room.objects.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		output.WriteString(fmt.Sprintf("    {W%s{x\r\n", obj.longDescription))
	}

	ch.Send(output.String())
}

func (room *Room) listOtherRoomCharactersToCharacter(ch *Character) {
	var output strings.Builder

	for iter := room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			output.WriteString(fmt.Sprintf("{G%s{x\r\n", rch.getLongDescription(ch)))
		}
	}

	ch.Send(output.String())
}

func (game *Game) NewRoom() *Room {
	room := &Room{game: game}

	return room
}

func (game *Game) LoadRoomIndex(index uint) (*Room, error) {
	room, ok := game.world[index]
	if ok {
		return room, nil
	}

	row := game.db.QueryRow(`
		SELECT
			id,
			zone_id,
			name,
			description,
			flags
		FROM
			rooms
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`, index)

	var zoneId int

	room = &Room{game: game}
	room.resets = NewLinkedList()
	room.objects = NewLinkedList()
	room.Characters = NewLinkedList()
	room.exit = make(map[uint]*Exit)
	err := row.Scan(&room.Id, &zoneId, &room.name, &room.description, &room.flags)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	for iter := game.Zones.Head; iter != nil; iter = iter.Next {
		zone := iter.Value.(*Zone)

		if zone.id == zoneId {
			room.zone = zone
		}
	}

	if room.zone == nil {
		return nil, errors.New("trying to instance a room without a zone")
	}

	game.world[room.Id] = room
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

/*
 * Utility method for the scripting engine to broadcast within a room using a filter fn
 */
func (room *Room) Broadcast(message string, filter goja.Callable) {
	var recipients []*Character = make([]*Character, 0)

	/* Collect characters in room for which filter(rch) == true */
	for iter := room.Characters.Head; iter != nil; iter = iter.Next {
		var result bool = false

		rch := iter.Value.(*Character)

		if filter != nil {
			val, err := filter(room.game.vm.ToValue(rch))
			if err != nil {
				log.Printf("Room.Broadcast failed: %v\r\n", err)
				break
			}

			result = val.ToBoolean()
		}

		if result || filter == nil {
			recipients = append(recipients, rch)
		}
	}

	/* Send message to gathered users */
	for _, rcpt := range recipients {
		rcpt.Send(message)
	}
}
