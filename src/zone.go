/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "log"

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
	id uint

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
