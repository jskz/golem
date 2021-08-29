/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"log"
)

type Zone struct {
	id uint

	name string
	low  uint
	high uint
}

const (
	ResetTypeMobile = 0
	ResetTypeObject = 1
	ResetTypeRoom   = 2
)

type Reset struct {
	id   uint
	zone *Zone
	room *Room

	resetType uint

	value0 int
	value1 int
	value2 int
	value3 int
}

func (game *Game) ResetRoom(room *Room) {
	for iter := room.resets.head; iter != nil; iter = iter.next {
		reset := iter.value.(*Reset)

		switch reset.resetType {
		case ResetTypeMobile:
			count := 0

			for rch := range room.characters {
				if rch.id == reset.value0 {
					count++
				}
			}

			if count >= reset.value2 {
				break
			}

			mobile, err := game.LoadMobileIndex(uint(reset.value0))
			if err != nil {
				log.Printf("Could not load mobile during reset: %v\r\n", err)
				break
			}

			if mobile != nil {
				room.characters[mobile] = true
				mobile.room = room
			}

		default:
			log.Printf("Reset of unknown type found for room.\r\n")
		}
	}
}

func (game *Game) ResetZone(zone *Zone) {
	for id := zone.low; id <= zone.high; id++ {
		room, err := game.LoadRoomIndex(id)
		if err != nil || room == nil {
			continue
		}

		game.ResetRoom(room)
	}
}

func (game *Game) LoadZones() error {
	log.Printf("Loading zones.\r\n")

	game.zones = make(map[*Zone]bool)

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			low,
			high
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
		zone := &Zone{}

		err := rows.Scan(&zone.id, &zone.name, &zone.low, &zone.high)
		if err != nil {
			log.Printf("Unable to scan zone row: %v.\r\n", err)
			continue
		}

		game.zones[zone] = true
	}

	log.Printf("Loaded %d zones from database.\r\n", len(game.zones))
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

		var zoneId uint
		var roomId uint
		var resetType string

		err := rows.Scan(&reset.id, &zoneId, &roomId, &resetType, &reset.value0, &reset.value1, &reset.value2, &reset.value3)
		if err != nil {
			log.Printf("Unable to scan reset row: %v.\r\n", err)
			continue
		}

		for zone := range game.zones {
			if zone.id == zoneId {
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

				reset.resetType, ok = resetEnumToUintType[resetType]
				if !ok {
					break
				}

				reset.zone = zone
				reset.room = room

				room.resets.Insert(reset)
				resetCount++
			}
		}
	}

	log.Printf("Loaded %d resets from database.\r\n", resetCount)
	return nil
}
