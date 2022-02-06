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
const RoomTradingPost = 8

/* Room flag types */
const (
	ROOM_PERSISTENT = 1
	ROOM_VIRTUAL    = 1 << 1
	ROOM_SAFE       = 1 << 2
	ROOM_DUNGEON    = 1 << 3
	ROOM_EVIL_AURA  = 1 << 4
	ROOM_PLANAR     = 1 << 5
	ROOM_DARK       = 1 << 6
)

var RoomFlagTable []Flag = []Flag{
	{Name: "persistent", Flag: ROOM_PERSISTENT},
	{Name: "virtual", Flag: ROOM_VIRTUAL},
	{Name: "safe", Flag: ROOM_SAFE},
	{Name: "dungeon", Flag: ROOM_DUNGEON},
	{Name: "evil_aura", Flag: ROOM_EVIL_AURA},
	{Name: "planar", Flag: ROOM_PLANAR},
	{Name: "dark", Flag: ROOM_DARK},
}

type Room struct {
	Game  *Game  `json:"game"`
	Plane *Plane `json:"plane"`
	Id    uint   `json:"id"`
	Zone  *Zone  `json:"zone"`

	script *Script

	Flags   int       `json:"flags"`
	Virtual bool      `json:"virtual"`
	Cell    *MazeCell `json:"cell"`
	X       int       `json:"x"`
	Y       int       `json:"y"`
	Z       int       `json:"z"`

	Name        string `json:"name"`
	Description string `json:"description"`

	Objects    *LinkedList `json:"objects"`
	Resets     *LinkedList `json:"resets"`
	Characters *LinkedList `json:"characters"`

	Exit map[uint]*Exit `json:"exit"`
}

func (room *Room) AddObject(obj *ObjectInstance) {
	room.Objects.Insert(obj)

	obj.Inside = nil
	obj.CarriedBy = nil
	obj.InRoom = room
}

func (room *Room) removeObject(obj *ObjectInstance) {
	room.Objects.Remove(obj)

	obj.InRoom = nil
	obj.CarriedBy = nil
	obj.Inside = nil
}

func FindRoomFlag(flag string) *Flag {
	for _, f := range RoomFlagTable {
		if strings.EqualFold(f.Name, flag) {
			return &f
		}
	}

	return nil
}

func (room *Room) ActiveLightSourcePresent() bool {
	if room.Flags&ROOM_DARK == 0 {
		return true
	}

	for iter := room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch.HasEquippedLightSource() {
			return true
		}
	}

	return false
}

func (room *Room) Visible(viewer *Character) bool {
	if viewer.Affected&AFFECT_BLINDNESS != 0 {
		return false
	}

	if room.Flags&ROOM_DARK != 0 && !room.ActiveLightSourcePresent() {
		return false
	}

	return true
}

func (room *Room) AddCharacter(ch *Character) {
	room.Characters.Insert(ch)
	ch.Room = room

	if room.Flags&ROOM_PLANAR != 0 && room.Plane != nil {
		ch.PlaneIndex = &Point{X: float64(room.X), Y: float64(room.Y), Value: ch}
		room.Plane.Map.Layers[room.Z].Atlas.CharacterTree.Insert(ch.PlaneIndex)

		for _, obs := range room.Plane.Map.Layers[room.Z].Observers {
			// If this obs does not contain the room, skip
			if !obs.Rect.Contains(float64(room.X), float64(room.Y)) {
				continue
			}

			var entering bool = true
			var previousRoom *Room = nil

			if len(ch.Trail) > 1 {
				previousRoom = ch.Trail[0]
			}

			if previousRoom != nil {
				if previousRoom.Flags&ROOM_PLANAR == 0 || previousRoom.Plane != room.Plane {
					entering = true
				} else if obs.Rect.Contains(float64(previousRoom.X), float64(previousRoom.Y)) && obs.Rect.Contains(float64(room.X), float64(room.Y)) {
					entering = false
				}
			}

			if entering {
				obs.OnEnterCallback(room.Game.vm.ToValue(ch))
			}
		}

		var previousRoom *Room = nil

		if len(ch.Trail) > 1 {
			previousRoom = ch.Trail[0]
		}

		if previousRoom != nil {
			if previousRoom.Flags&ROOM_PLANAR != 0 {
				for _, obs := range previousRoom.Plane.Map.Layers[previousRoom.Z].Observers {
					var leaving bool = false

					if previousRoom.Flags&ROOM_PLANAR == 0 || previousRoom.Plane != room.Plane {
						leaving = true
					} else if obs.Rect.Contains(float64(previousRoom.X), float64(previousRoom.Y)) && !obs.Rect.Contains(float64(room.X), float64(room.Y)) {
						leaving = true
					}

					if leaving {
						obs.OnLeaveCallback(room.Game.vm.ToValue(ch))
					}
				}
			}
		}
	}

	trail := make([]*Room, 0)
	trail = append(trail, room)
	trail = append(trail, ch.Trail...)

	if len(trail) > 5 {
		trail = trail[0:5]
	}

	ch.Trail = trail
}

func (room *Room) removeCharacter(ch *Character) {
	room.Characters.Remove(ch)

	// If the origin room was planar, remove this character from its atlas' character lookup quadtree
	if room.Flags&ROOM_PLANAR != 0 && room.Plane != nil {
		room.Plane.Map.Layers[room.Z].Atlas.CharacterTree.Remove(ch.PlaneIndex)
	}

	ch.Room = nil
}

func (room *Room) listObjectsToCharacter(ch *Character) {
	ch.listObjects(room.Objects, true, false)
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
	room := &Room{Game: game}

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

	room = &Room{Game: game}
	room.Resets = NewLinkedList()
	room.Objects = NewLinkedList()
	room.Characters = NewLinkedList()
	room.Exit = make(map[uint]*Exit)
	err := row.Scan(&room.Id, &zoneId, &room.Name, &room.Description, &room.Flags)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	for iter := game.Zones.Head; iter != nil; iter = iter.Next {
		zone := iter.Value.(*Zone)

		if zone.Id == zoneId {
			room.Zone = zone
		}
	}

	if room.Zone == nil {
		return nil, errors.New("trying to instance a room without a zone")
	}

	game.world[room.Id] = room
	return room, nil
}

func (room *Room) IsEqual(oroom *Room) bool {
	return (oroom == room || ((room.Flags&ROOM_PLANAR != 0 && oroom.Flags&ROOM_PLANAR != 0) &&
		(room.Plane == oroom.Plane) &&
		(room.X == oroom.X && room.Y == oroom.Y && room.Z == oroom.Z)))
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

		rows.Scan(&exit.Id, &roomId, &toRoomId, &exit.Direction, &exit.Flags)
		exit.To, err = game.LoadRoomIndex(uint(toRoomId))
		if err != nil {
			continue
		}

		exit.Room, err = game.LoadRoomIndex(uint(roomId))
		if err != nil {
			continue
		}

		exit.Room.Exit[exit.Direction] = exit
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
			val, err := filter(room.Game.vm.ToValue(rch))
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
