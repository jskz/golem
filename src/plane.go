/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"errors"
	"log"
	"sync"
)

type Plane struct {
	Game       *Game   `json:"game"`
	Zone       *Zone   `json:"zone"`
	Id         int     `json:"id"`
	Flags      int     `json:"flags"`
	Name       string  `json:"name"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	Depth      int     `json:"depth"`
	PlaneType  string  `json:"planeType"`
	SourceType string  `json:"sourceType"`
	Scripts    *Script `json:"scripts"`

	Maze    *MazeGrid   `json:"maze"`
	Portals *LinkedList `json:"portals"`
}

type Portal struct {
	Id         int    `json:"id"`
	PortalType string `json:"portalType"`
	Room       *Room  `json:"room"`
	Plane      *Plane `json:"plane"`
}

/* plane flags */
const (
	PLANE_INITIALIZED = 0
)

/* plane_type ENUM values */
const (
	PlaneTypeVoid       = "void"
	PlaneTypeMaze       = "maze"
	PlaneTypeWilderness = "wilderness"
)

/* source_type ENUM values */
const (
	SourceTypeVoid       = "void"
	SourceTypeBlob       = "blob"
	SourceTypeProcedural = "procedural"
)

/* portal_type ENUM values */
const (
	PortalTypeFixed      = "fixed"
	PortalTypeProcedural = "procedural"
)

func (plane *Plane) generate() error {
	game := plane.Game

	switch plane.PlaneType {
	case PlaneTypeMaze:
		switch plane.SourceType {
		case SourceTypeProcedural:
			log.Printf("Generating maze with dimensions %dx%dx%d for plane %d.\r\n", plane.Width, plane.Height, plane.Depth, plane.Id)

			/*
			 * TODO: copied from previous maze test, allow a plane entrance coordinate or location on the model?
			 *
			 * Hardcode an exit from limbo into the first floor of the test dungeon
			 */
			limbo, err := game.LoadRoomIndex(RoomLimbo)
			if err != nil {
				log.Println(err)
			}

			/* Exit will be self-referential and locked until the maze is done generating */
			limbo.exit[DirectionDown] = &Exit{
				id:        0,
				direction: DirectionDown,
				to:        limbo,
				flags:     EXIT_IS_DOOR | EXIT_CLOSED | EXIT_LOCKED,
			}

			go func() {
				var dungeon *Dungeon

				wg := sync.WaitGroup{}
				wg.Add(1)

				go func() {
					dungeon = game.GenerateDungeon(plane.Depth, plane.Width, plane.Height)
					wg.Done()
				}()

				wg.Wait()

				if dungeon == nil || len(dungeon.floors) < 1 {
					log.Printf("Dungeon generation attempt aborting.\r\n")
					return
				}

				maze := dungeon.floors[0]

				/* Unlock the entrance */
				limbo.exit[DirectionDown].to = maze.grid[maze.entryX][maze.entryY].room
				limbo.exit[DirectionDown].flags &= ^EXIT_LOCKED

				maze.grid[maze.entryX][maze.entryY].room.exit[DirectionUp] = &Exit{
					id:        0,
					direction: DirectionUp,
					to:        limbo,
					flags:     EXIT_IS_DOOR | EXIT_CLOSED,
				}
			}()
		default:
			return errors.New("unimplemented maze source type")
		}
	default:
		return errors.New("unimplemented plane type")
	}

	return nil
}

func (game *Game) FindPlaneByID(id int) *Plane {
	for iter := game.Planes.Head; iter != nil; iter = iter.Next {
		plane := iter.Value.(*Plane)

		if plane.Id == id {
			return plane
		}
	}

	return nil
}

func (game *Game) LoadPlanes() error {
	log.Printf("Loading planes.\r\n")

	rows, err := game.db.Query(`
		SELECT
			id,
			zone_id,
			name,
			width,
			height,
			depth,
			plane_type,
			source_type
		FROM
			planes
	`)
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		plane := &Plane{Game: game}
		plane.Portals = NewLinkedList()
		plane.Flags = 0

		var zoneId int = 0

		err := rows.Scan(&plane.Id, &zoneId, &plane.Name, &plane.Width, &plane.Height, &plane.Depth, &plane.PlaneType, &plane.SourceType)
		if err != nil {
			log.Printf("Unable to scan plane: %v.\r\n", err)
			return err
		}

		for iter := game.Zones.Head; iter != nil; iter = iter.Next {
			zone := iter.Value.(*Zone)

			if zone.id == zoneId {
				plane.Zone = zone
			}
		}

		if plane.Zone == nil {
			return errors.New("trying to load plane with a bad zone")
		}

		game.Planes.Insert(plane)
	}

	log.Printf("Loaded %d planes from database.\r\n", game.Planes.Count)
	return nil
}
