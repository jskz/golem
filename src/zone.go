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
	Id   uint
	Zone *Zone
	Room *Room

	ResetType uint

	Value0 int
	Value1 int
	Value2 int
	Value3 int
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

				room.addObject(obj)
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
