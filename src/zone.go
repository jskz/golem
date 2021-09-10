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
)

type Zone struct {
	id uint

	name string
	low  uint
	high uint

	resetMessage string
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
		case ResetTypeObject:
			count := 0

			for iter := room.objects.head; iter != nil; iter = iter.next {
				obj := iter.value.(*ObjectInstance)

				if obj.parentId == uint(reset.value0) {
					count++
				}
			}

			if count >= reset.value2 {
				break
			}

			/* Create a new object instance and place it in the room */
			objIndex, err := game.LoadObjectIndex(uint(reset.value0))
			if err != nil {
				log.Printf("Failed to load object for reset: %v\r\n", err)
				continue
			}

			if objIndex != nil {
				obj := &ObjectInstance{
					parentId:         objIndex.id,
					name:             objIndex.name,
					shortDescription: objIndex.shortDescription,
					longDescription:  objIndex.longDescription,
					description:      objIndex.description,
					value0:           objIndex.value0,
					value1:           objIndex.value1,
					value2:           objIndex.value2,
					value3:           objIndex.value3,
				}

				room.addObject(obj)
			}

		case ResetTypeMobile:
			count := 0

			for iter := room.characters.head; iter != nil; iter = iter.next {
				rch := iter.value.(*Character)

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
				room.addCharacter(mobile)

				game.characters.Insert(mobile)
			}

		default:
			log.Printf("Reset of unknown type found for room.\r\n")
		}
	}

	for iter := room.characters.head; iter != nil; iter = iter.next {
		character := iter.value.(*Character)
		character.onZoneUpdate()

		if room.zone.resetMessage != "" && character.flags&CHAR_IS_PLAYER != 0 {
			character.Send(fmt.Sprintf("\r\n{x%s{x\r\n", room.zone.resetMessage))
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

	game.zones = NewLinkedList()

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			low,
			high,
			reset_message
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

		err := rows.Scan(&zone.id, &zone.name, &zone.low, &zone.high, &zone.resetMessage)
		if err != nil {
			log.Printf("Unable to scan zone row: %v.\r\n", err)
			continue
		}

		game.zones.Insert(zone)
	}

	log.Printf("Loaded %d zones from database.\r\n", game.zones.count)
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

		for iter := game.zones.head; iter != nil; iter = iter.next {
			zone := iter.value.(*Zone)

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
