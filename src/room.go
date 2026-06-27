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

	Objects    *LinkedList[*ObjectInstance] `json:"objects"`
	Resets     *LinkedList[*Reset]          `json:"resets"`
	Characters *LinkedList[*Character]      `json:"characters"`

	Exit map[uint]*Exit `json:"exit"`
}

func (room *Room) AddObject(obj *ObjectInstance) {
	room.Objects.Insert(obj)

	obj.Inside = nil
	obj.CarriedBy = nil
	obj.InRoom = room
	obj.WearLocation = -1
}

func (room *Room) removeObject(obj *ObjectInstance) {
	room.clearFurnitureUsers(obj)
	room.Objects.Remove(obj)

	obj.InRoom = nil
	obj.CarriedBy = nil
	obj.Inside = nil
}

func (room *Room) clearFurnitureUsers(obj *ObjectInstance) {
	if room == nil || room.Characters == nil || obj == nil {
		return
	}

	for iter := room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value
		if rch.Furniture == obj {
			rch.Furniture = nil
		}
	}
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
		rch := iter.Value

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

func (room *Room) planarLayer() (*MapGrid, bool) {
	if room == nil || room.Flags&ROOM_PLANAR == 0 || room.Plane == nil || room.Plane.Map == nil {
		return nil, false
	}

	if room.Z < 0 || room.Z >= len(room.Plane.Map.Layers) {
		return nil, false
	}

	layer := room.Plane.Map.Layers[room.Z]
	if layer == nil {
		return nil, false
	}

	return layer, true
}

func (room *Room) samePlanarLayer(other *Room) bool {
	if room == nil || other == nil {
		return false
	}

	return room.Flags&ROOM_PLANAR != 0 &&
		other.Flags&ROOM_PLANAR != 0 &&
		room.Plane != nil &&
		room.Plane == other.Plane &&
		room.Z == other.Z
}

func (room *Room) notifyPlaneEnter(ch *Character, previousRoom *Room) {
	layer, ok := room.planarLayer()
	if !ok {
		return
	}

	for _, obs := range layer.Observers {
		if obs.Rect == nil || obs.OnEnterCallback == nil {
			continue
		}

		if !obs.Rect.Contains(float64(room.X), float64(room.Y)) {
			continue
		}

		if room.samePlanarLayer(previousRoom) && obs.Rect.Contains(float64(previousRoom.X), float64(previousRoom.Y)) {
			continue
		}

		obs.OnEnterCallback(room.Game.vm.ToValue(ch))
	}
}

func (room *Room) notifyPlaneLeave(ch *Character, destination *Room) {
	layer, ok := room.planarLayer()
	if !ok {
		return
	}

	for _, obs := range layer.Observers {
		if obs.Rect == nil || obs.OnLeaveCallback == nil {
			continue
		}

		if !obs.Rect.Contains(float64(room.X), float64(room.Y)) {
			continue
		}

		if room.samePlanarLayer(destination) && obs.Rect.Contains(float64(destination.X), float64(destination.Y)) {
			continue
		}

		obs.OnLeaveCallback(room.Game.vm.ToValue(ch))
	}
}

func (room *Room) AddCharacter(ch *Character) {
	previousRoom := ch.moveOrigin
	if previousRoom == nil && len(ch.Trail) > 0 {
		previousRoom = ch.Trail[0]
	}
	ch.moveOrigin = nil

	room.Characters.Insert(ch)
	ch.Room = room

	if layer, ok := room.planarLayer(); ok {
		ch.PlaneIndex = &Point{X: float64(room.X), Y: float64(room.Y), Value: ch}
		if layer.Atlas != nil && layer.Atlas.CharacterTree != nil {
			layer.Atlas.CharacterTree.Insert(ch.PlaneIndex)
		}

		room.notifyPlaneEnter(ch, previousRoom)
	}

	trail := make([]*Room, 0)
	trail = append(trail, room)
	trail = append(trail, ch.Trail...)

	if len(trail) > 5 {
		trail = trail[0:5]
	}

	ch.Trail = trail
}

func (room *Room) moveCharacter(ch *Character, destination *Room) {
	if destination == nil {
		room.removeCharacter(ch)
		return
	}

	room.removeCharacterForMove(ch, destination)
	destination.AddCharacter(ch)
}

func (room *Room) removeCharacter(ch *Character) {
	room.removeCharacterForMove(ch, nil)
	ch.moveOrigin = nil
}

func (room *Room) removeCharacterForMove(ch *Character, destination *Room) {
	room.Characters.Remove(ch)
	ch.moveOrigin = room
	ch.Furniture = nil

	// If the origin room was planar, remove this character from its atlas' character lookup quadtree
	if layer, ok := room.planarLayer(); ok {
		if ch.PlaneIndex != nil && layer.Atlas != nil && layer.Atlas.CharacterTree != nil {
			layer.Atlas.CharacterTree.Remove(ch.PlaneIndex)
		}
		ch.PlaneIndex = nil
		room.notifyPlaneLeave(ch, destination)
	}

	ch.Room = nil
}

func (room *Room) listObjectsToCharacter(ch *Character) {
	ch.listObjects(room.Objects, true, false)
}

func (room *Room) listOtherRoomCharactersToCharacter(ch *Character) {
	var output strings.Builder

	for iter := room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value

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
	room.Resets = NewLinkedList[*Reset]()
	room.Objects = NewLinkedList[*ObjectInstance]()
	room.Characters = NewLinkedList[*Character]()
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

	type exitRecord struct {
		exit     *Exit
		roomId   int
		toRoomId int
	}

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

	exitRecords := make([]exitRecord, 0)

	for rows.Next() {
		exit := &Exit{}

		var roomId int
		var toRoomId int

		err = rows.Scan(&exit.Id, &roomId, &toRoomId, &exit.Direction, &exit.Flags)
		if err != nil {
			return err
		}

		exitRecords = append(exitRecords, exitRecord{
			exit:     exit,
			roomId:   roomId,
			toRoomId: toRoomId,
		})
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if err := rows.Close(); err != nil {
		return err
	}

	for _, record := range exitRecords {
		exit := record.exit

		exit.To, err = game.LoadRoomIndex(uint(record.toRoomId))
		if err != nil {
			return fmt.Errorf("failed to load destination room %d for exit %d: %w", record.toRoomId, exit.Id, err)
		}
		if exit.To == nil {
			return fmt.Errorf("exit %d references missing or deleted destination room %d", exit.Id, record.toRoomId)
		}

		exit.Room, err = game.LoadRoomIndex(uint(record.roomId))
		if err != nil {
			return fmt.Errorf("failed to load source room %d for exit %d: %w", record.roomId, exit.Id, err)
		}
		if exit.Room == nil {
			return fmt.Errorf("exit %d references missing or deleted source room %d", exit.Id, record.roomId)
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

		rch := iter.Value

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
