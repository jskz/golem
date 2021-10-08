/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "log"

type Plane struct {
	Id int `json:"id"`
}

/* Plane type ENUM values */
const (
	PlaneTypeVoid       = "void"
	PlaneTypeMaze       = "maze"
	PlaneTypeWilderness = "wilderness"
)

func (game *Game) LoadPlanes() error {
	log.Printf("Loading planes.\r\n")

	rows, err := game.db.Query(`
		SELECT
			id,
			name
		FROM
			planes
	`)
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		plane := &Plane{}
		err := rows.Scan(&plane.Id)
		if err != nil {
			log.Printf("Unable to scan plane: %v.\r\n", err)
			return err
		}

		game.Planes.Insert(plane)
	}

	log.Printf("Loaded %d planes from database.\r\n", game.Planes.Count)
	return nil
}
