/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"context"
	"fmt"
	"log"
	"time"
)

type Zone struct {
	Game           *Game  `json:"game"`
	Id             int    `json:"id"`
	Name           string `json:"name"`
	WhoDescription string `json:"whoDescription"`
	Low            uint   `json:"low"`
	High           uint   `json:"high"`

	ResetMessage   string    `json:"resetMessage"`
	ResetFrequency int       `json:"resetFrequency"`
	LastReset      time.Time `json:"lastReset"`
}

const (
	ResetTypeMobile = 0
	ResetTypeObject = 1
	ResetTypeRoom   = 2
)

type Reset struct {
	Id   uint  `json:"id"`
	Zone *Zone `json:"zone"`
	Room *Room `json:"room"`

	ResetType uint `json:"resetType"`

	Value0 int `json:"value0"`
	Value1 int `json:"value1"`
	Value2 int `json:"value2"`
	Value3 int `json:"value3"`
}

func (room *Room) CreateReset(resetType uint, v0, v1, v2, v3 int) *Reset {
	if room.Flags&ROOM_VIRTUAL != 0 || room.Flags&ROOM_PLANAR != 0 || room.Zone == nil || room.Zone.Id == 0 {
		return nil
	}

	var typeString string = ""
	switch resetType {
	case ResetTypeMobile:
		typeString = "mobile"
	case ResetTypeObject:
		typeString = "object"
	default:
		return nil
	}

	res, err := room.Game.db.Exec(`
	INSERT INTO
		resets(zone_id, room_id, type, value_1, value_2, value_3, value_4)
	VALUES
		(?, ?, ?, ?, ?, ?, ?)
	`, room.Zone.Id, room.Id, typeString, v0, v1, v2, v3)
	if err != nil {
		return nil
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return nil
	}

	reset := &Reset{
		Id:   uint(lastInsertId),
		Zone: room.Zone,
		Room: room,

		ResetType: resetType,

		Value0: v0,
		Value1: v1,
		Value2: v2,
		Value3: v3,
	}

	if err != nil {
		return nil
	}

	return reset
}

func (game *Game) CreateZone() *Zone {
	zone := &Zone{}
	zone.Name = "Untitled Zone"
	zone.WhoDescription = "Void"
	zone.Low = 0
	zone.High = 0
	zone.ResetMessage = "Tick-tock."
	zone.ResetFrequency = 15

	res, err := game.db.Exec(`
	INSERT INTO
		zones(name, who_description, low, high, reset_message, reset_frequency)
	VALUES
		(?, ?, ?, ?, ?, ?)
	`, zone.Name, zone.WhoDescription, zone.Low, zone.High, zone.ResetMessage, zone.ResetFrequency)
	if err != nil {
		return nil
	}

	lastInsertId, err := res.LastInsertId()
	if err != nil {
		return nil
	}

	zone.Id = int(lastInsertId)
	return zone
}

func (game *Game) ResetRoom(room *Room) {
	for iter := room.Resets.Head; iter != nil; iter = iter.Next {
		reset := iter.Value.(*Reset)

		switch reset.ResetType {
		case ResetTypeObject:
			count := 0

			for iter := room.Objects.Head; iter != nil; iter = iter.Next {
				obj := iter.Value.(*ObjectInstance)

				if obj.ParentId == uint(reset.Value0) {
					count++
				}
			}

			if count >= reset.Value2 {
				break
			}

			/* Create a new object instance and place it in the room */
			objIndex, err := game.LoadObjectIndex(uint(reset.Value0))
			if err != nil {
				log.Printf("Failed to load object for reset: %v\r\n", err)
				continue
			}

			if objIndex != nil {
				obj := &ObjectInstance{
					Game:             game,
					ParentId:         objIndex.Id,
					Contents:         NewLinkedList(),
					Inside:           nil,
					CarriedBy:        nil,
					Flags:            objIndex.Flags,
					Name:             objIndex.Name,
					ShortDescription: objIndex.ShortDescription,
					LongDescription:  objIndex.LongDescription,
					Description:      objIndex.Description,
					ItemType:         objIndex.ItemType,
					Value0:           objIndex.Value0,
					Value1:           objIndex.Value1,
					Value2:           objIndex.Value2,
					Value3:           objIndex.Value3,
					CreatedAt:        time.Now(),
					WearLocation:     -1,
				}

				room.AddObject(obj)
				game.Objects.Insert(obj)
			}

		case ResetTypeMobile:
			count := 0

			for iter := room.Characters.Head; iter != nil; iter = iter.Next {
				rch := iter.Value.(*Character)

				if rch.Flags&CHAR_IS_PLAYER == 0 && rch.Id == reset.Value0 {
					count++
				}
			}

			if count >= reset.Value2 {
				break
			}

			mobile, err := game.LoadMobileIndex(uint(reset.Value0))
			if err != nil {
				log.Printf("Could not load mobile during reset: %v\r\n", err)
				break
			}

			if mobile != nil {
				room.AddCharacter(mobile)
				game.Characters.Insert(mobile)
			}

		default:
			log.Printf("Reset of unknown type found for room.\r\n")
		}
	}

	for iter := room.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)

		if room.Zone.ResetMessage != "" && character.Flags&CHAR_IS_PLAYER != 0 {
			character.Send(fmt.Sprintf("\r\n{x%s{x\r\n", room.Zone.ResetMessage))
		}
	}
}

func (game *Game) ValidZoneRange(low uint, high uint) bool {
	if low <= 0 || high <= 0 || high < low || low >= 2147483647 || high >= 2147483647 {
		return false
	}

	for zoneIter := game.Zones.Head; zoneIter != nil; zoneIter = zoneIter.Next {
		zone := zoneIter.Value.(*Zone)

		if low <= zone.High && zone.Low <= high {
			return false
		}
	}

	return true
}

func (game *Game) ResetZone(zone *Zone) {
	for id := zone.Low; id <= zone.High; id++ {
		room, err := game.LoadRoomIndex(id)

		if err != nil || room == nil {
			continue
		}

		game.ResetRoom(room)
	}

	zone.LastReset = time.Now()
}

// If possible, create a new room record for this zone and return its instance
func (zone *Zone) CreateRoom() (*Room, error) {
	ctx := context.Background()

	tx, err := zone.Game.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	var availableId uint = 0

	row := tx.QueryRow(`
	/*
	* Zones are not intended to be larger than a few hundred rooms at most and will never
	* exceed the recursion depth limit; first, create the sequence of numbers in the low-high
	* range for this zone's specified range.
	*/
	WITH RECURSIVE room_ids AS (
		SELECT (
			SELECT low FROM zones WHERE zones.id = ?
		) AS value
		UNION ALL
		SELECT value + 1 AS value
		FROM room_ids
		WHERE room_ids.value < (
			SELECT high FROM zones WHERE zones.id = ?
		)
	)
	SELECT
		value
	FROM
		room_ids
	WHERE
		value
	/* Exclude the set of this zone's room IDs */
	NOT IN (
		SELECT
			rooms.id
		FROM
			rooms
		INNER JOIN
			zones ON zones.id = rooms.zone_id
	);`, zone.Id, zone.Id)

	err = row.Scan(&availableId)
	if err != nil {
		tx.Rollback()

		return nil, err
	}

	room := zone.Game.NewRoom()
	room.Id = availableId
	room.Zone = zone
	room.Name = "Empty Room"
	room.Flags = 0
	room.Description = "This room-in-development needs a description!"
	room.Exit = make(map[uint]*Exit)
	room.Characters = NewLinkedList()
	room.Objects = NewLinkedList()
	room.Resets = NewLinkedList()

	_, err = tx.Exec(`
		INSERT INTO
			rooms(id, zone_id, flags, name, description)
		VALUES
			(?, ?, ?, ?, ?)
	`, availableId, zone.Id, room.Flags, room.Name, room.Description)
	if err != nil {
		tx.Rollback()

		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	zone.Game.world[availableId] = room

	return room, nil
}

func (zone *Zone) FindAvailableRoomID() (int, error) {
	var availableId int = 0

	row := zone.Game.db.QueryRow(`
	/*
	* Zones are not intended to be larger than a few hundred rooms at most and will never
	* exceed the recursion depth limit; first, create the sequence of numbers in the low-high
	* range for this zone's specified range.
	*/
	WITH RECURSIVE room_ids AS (
		SELECT (
			SELECT low FROM zones WHERE zones.id = ?
		) AS value
		UNION ALL
		SELECT value + 1 AS value
		FROM room_ids
		WHERE room_ids.value < (
			SELECT high FROM zones WHERE zones.id = ?
		)
	)
	SELECT
		value
	FROM
		room_ids
	WHERE
		value
	/* Exclude the set of this zone's room IDs */
	NOT IN (
		SELECT
			rooms.id
		FROM
			rooms
		INNER JOIN
			zones ON zones.id = rooms.zone_id
	);`, zone.Id, zone.Id)

	err := row.Scan(&availableId)
	if err != nil {
		return 0, err
	}

	return availableId, nil
}

func (game *Game) LoadZones() error {
	log.Printf("Loading zones.\r\n")

	game.Zones = NewLinkedList()

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			who_description,
			low,
			high,
			reset_message,
			reset_frequency
		FROM
			zones
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		zone := &Zone{Game: game}

		err := rows.Scan(&zone.Id, &zone.Name, &zone.WhoDescription, &zone.Low, &zone.High, &zone.ResetMessage, &zone.ResetFrequency)
		if err != nil {
			log.Printf("Unable to scan zone row: %v.\r\n", err)
			continue
		}

		game.Zones.Insert(zone)
	}

	log.Printf("Loaded %d zones from database.\r\n", game.Zones.Count)
	return nil
}

func (game *Game) LoadResets() error {
	var resetCount int = 0

	log.Printf("Loading resets.\r\n")

	rows, err := game.db.Query(`
		SELECT
			id,
			zone_id,
			room_id,
			type,
			value_1,
			value_2,
			value_3,
			value_4
		FROM
			resets
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		reset := &Reset{}

		var zoneId int
		var roomId uint
		var resetType string

		err := rows.Scan(&reset.Id, &zoneId, &roomId, &resetType, &reset.Value0, &reset.Value1, &reset.Value2, &reset.Value3)
		if err != nil {
			log.Printf("Unable to scan reset row: %v.\r\n", err)
			continue
		}

		for iter := game.Zones.Head; iter != nil; iter = iter.Next {
			zone := iter.Value.(*Zone)

			if zone.Id == zoneId {
				room, err := game.LoadRoomIndex(roomId)
				if err != nil {
					return err
				}

				//`type` ENUM('mobile', 'room', 'object')
				var resetEnumToUintType = map[string]uint{
					"mobile": ResetTypeMobile,
					"room":   ResetTypeRoom,
					"object": ResetTypeObject,
				}

				var ok bool

				reset.ResetType, ok = resetEnumToUintType[resetType]
				if !ok {
					break
				}

				reset.Zone = zone
				reset.Room = room

				room.Resets.Insert(reset)

				resetCount++
			}
		}
	}

	log.Printf("Loaded %d resets from database.\r\n", resetCount)
	return nil
}

func (game *Game) FindZoneByID(id int) *Zone {
	for zoneIter := game.Zones.Head; zoneIter != nil; zoneIter = zoneIter.Next {
		zone := zoneIter.Value.(*Zone)

		if zone.Id == id {
			return zone
		}
	}

	return nil
}
