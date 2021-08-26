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

func (game *Game) LoadZones() error {
	log.Printf("Loading zones.\r\n")

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
