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
	"fmt"
	"log"
	"strings"
	"sync"
)

type Plane struct {
	Game       *Game    `json:"game"`
	Zone       *Zone    `json:"zone"`
	Dungeon    *Dungeon `json:"dungeon"`
	Id         int      `json:"id"`
	Flags      int      `json:"flags"`
	Name       string   `json:"name"`
	Width      int      `json:"width"`
	Height     int      `json:"height"`
	Depth      int      `json:"depth"`
	PlaneType  string   `json:"planeType"`
	SourceType string   `json:"sourceType"`
	Scripts    *Script  `json:"scripts"`

	Map     *Map        `json:"map"`
	Maze    *MazeGrid   `json:"maze"`
	Portals *LinkedList `json:"portals"`
}

type MapGrid struct {
	Terrain [][]int `json:"terrain"`
	Atlas   *Atlas  `json:"atlas"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
}

type Map struct {
	Layers []*MapGrid `json:"layers"`
}

// Atlas will be a collection of quadtrees for a plane providing spacial indices to quickly lookup:
// temporary instances of in-memory rooms, characters, objects, misc game entities within that plane;
// these are unused interface stubs until quadtree branch is ready
type Atlas struct {
	Characters interface{}
	Objects    interface{}
	Rooms      interface{}

	// portals, scripts, exits?
}

type Portal struct {
	Id         int    `json:"id"`
	PortalType string `json:"portalType"`
	Room       *Room  `json:"room"`
	Plane      *Plane `json:"plane"`
}

/* plane flags */
const (
	PLANE_NONE        = 0
	PLANE_INITIALIZED = 1
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

func NewAtlas() *Atlas {
	return &Atlas{
		// init quadtrees
	}
}

// Fill the source_value field for this plane with an appropriately sized binary blob of zeroes
func (plane *Plane) InitializeBlob() ([]byte, int, error) {
	log.Printf("Initializing new blob for plane %d.\r\n", plane.Id)

	plane.Map = &Map{
		Layers: make([]*MapGrid, 0),
	}

	var bytes []byte = make([]byte, plane.Depth*plane.Width*plane.Height)

	for z := 0; z < plane.Depth; z++ {
		grid := &MapGrid{}
		grid.Terrain = make([][]int, plane.Height)

		for y := 0; y < plane.Height; y++ {
			grid.Terrain[y] = make([]int, plane.Width)

			for x := 0; x < plane.Width; x++ {
				grid.Terrain[y][x] = 0
				bytes = append(bytes, 0)
			}
		}

		plane.Map.Layers = append(plane.Map.Layers, grid)
	}

	_, err := plane.Game.db.Exec(`
		UPDATE
			planes
		SET
			source_value = ?
		WHERE
			id = ?
	`, bytes, plane.Id)
	if err != nil {
		return nil, 0, err
	}

	return bytes, plane.Depth * plane.Width * plane.Height, nil
}

func (ch *Character) CreatePlaneMap() string {
	if ch.Room == nil || ch.Room.Plane == nil {
		return "Error retrieving plane map\r\n"
	}

	var buf strings.Builder

	rect := ch.Room.Plane.GetTerrainRect(0, 0, 0, 26, 14)
	if len(rect) != 0 {
		lastColour := ""

		for cY := 0; cY < 14; cY++ {
			for cX := 0; cX < 26; cX++ {
				if ch.Room.X == cX && cY == ch.Room.Y {
					buf.WriteString("{Y@")
					lastColour = "{Y"
					continue
				}

				_, ok := TerrainTable[rect[cY][cX]]
				if !ok {
					buf.WriteString("{x ")
					lastColour = ""
					continue
				}

				if lastColour == "" || lastColour != TerrainTable[rect[cY][cX]].GlyphColour {
					lastColour = TerrainTable[rect[cY][cX]].GlyphColour
					buf.WriteString(lastColour)
				}

				buf.WriteString(TerrainTable[rect[cY][cX]].MapGlyph)
			}

			buf.WriteString("\r\n")
		}

		return buf.String()
	}

	return fmt.Sprintf("Erroneous plane ID %d map from position (%d, %d, %d)\r\n", ch.Room.Plane.Id, ch.Room.X, ch.Room.Y, ch.Room.Z)
}

func (plane *Plane) generate() error {
	game := plane.Game

	if plane.Scripts != nil {
		plane.Scripts.tryEvaluate("onGenerate", plane.Game.vm.ToValue(game), plane.Game.vm.ToValue(plane))
	}

	switch plane.PlaneType {
	case PlaneTypeWilderness:
		switch plane.SourceType {
		case SourceTypeBlob:
			log.Printf("Initializing a %dx%dx%d wilderness zone from a data blob for plane %d.\r\n", plane.Width, plane.Height, plane.Depth, plane.Id)

			row := game.db.QueryRow(`
				SELECT
					(CASE
						WHEN source_value IS NULL THEN -1
						ELSE LENGTH(source_value)
					END),
					(CASE
						WHEN source_value IS NULL THEN -1
						ELSE source_value
					END)
				FROM
					planes
				WHERE
					id = ?`,
				plane.Id)

			var blobSize int = 0
			var blob []byte = make([]byte, plane.Depth*plane.Width*plane.Height)

			err := row.Scan(&blobSize, &blob)
			if err != nil {
				return err
			}

			if blobSize == -1 {
				blob, blobSize, err = plane.InitializeBlob()
				if err != nil {
					log.Printf("Plane %d remaining uninitialized after load with a NULL blob.\r\n", plane.Id)
					return nil
				}
			}

			planeMap := &Map{
				Layers: make([]*MapGrid, 0),
			}

			for z := 0; z < plane.Depth; z++ {
				grid := &MapGrid{}
				grid.Terrain = make([][]int, plane.Height)

				for y := 0; y < plane.Height; y++ {
					grid.Terrain[y] = make([]int, plane.Width)

					for x := 0; x < plane.Width; x++ {
						grid.Terrain[y][x] = int(blob[(z*(y*plane.Height))+x])
					}
				}

				planeMap.Layers = append(planeMap.Layers, grid)
			}

			log.Printf("Plane %d initialized from %d byte blob.\r\n", plane.Id, blobSize)

			plane.Flags |= PLANE_INITIALIZED
			plane.Map = planeMap
		}
	case PlaneTypeMaze:
		switch plane.SourceType {
		case SourceTypeProcedural:
			log.Printf("Generating maze with dimensions %dx%dx%d for plane %d.\r\n", plane.Width, plane.Height, plane.Depth, plane.Id)

			/*
				 * TODO: copied from previous maze test, allow a plane entrance coordinate or location on the model?
				 *
				 * Hardcode an exit from limbo into the first floor of the test dungeon
				limbo, err := game.LoadRoomIndex(RoomLimbo)
				if err != nil {
					log.Println(err)
				}
			*/

			/* Exit will be self-referential and locked until the maze is done generating */

			/*
				limbo.exit[DirectionDown] = &Exit{
					id:        0,
					direction: DirectionDown,
					to:        limbo,
					flags:     EXIT_IS_DOOR | EXIT_CLOSED | EXIT_LOCKED,
				}
			*/

			go func() {
				var dungeon *Dungeon

				wg := sync.WaitGroup{}
				wg.Add(1)

				go func() {
					dungeon = game.GenerateDungeon(plane.Depth, plane.Width, plane.Height)

					wg.Done()
				}()

				wg.Wait()

				if dungeon == nil || len(dungeon.Floors) < 1 {
					log.Printf("Dungeon generation attempt aborting.\r\n")
					return
				}

				plane.Dungeon = dungeon
				game.planeGenerationCompleted <- plane.Id

				/* Unlock the entrance */
				/*
					maze := dungeon.Floors[0]
						limbo.exit[DirectionDown].to = maze.grid[maze.entryX][maze.entryY].room
						limbo.exit[DirectionDown].flags &= ^EXIT_LOCKED

						maze.grid[maze.entryX][maze.entryY].room.exit[DirectionUp] = &Exit{
							id:        0,
							direction: DirectionUp,
							to:        limbo,
							flags:     EXIT_IS_DOOR | EXIT_CLOSED,
						}
				*/
			}()
		default:
			return errors.New("unimplemented maze source type")
		}
	default:
		return errors.New("unimplemented plane type")
	}

	return nil
}

func (plane *Plane) MaterializeRoom(x int, y int, z int) *Room {
	if plane.PlaneType == "dungeon" {
		return plane.Dungeon.Floors[z].Grid[y][x].Room
	}

	room := plane.Game.NewRoom()
	room.Plane = plane
	room.Id = 0
	room.Name = "Holodeck"
	room.Description = "If you are seeing this message, something has gone wrong."
	room.Flags = ROOM_VIRTUAL | ROOM_PLANAR
	room.Characters = NewLinkedList()
	room.Objects = NewLinkedList()
	room.Exit = make(map[uint]*Exit)
	room.X = x
	room.Y = y
	room.Z = y

	return room
}

func (plane *Plane) GetTerrainRect(x int, y int, z int, w int, h int) [][]int {
	var rectWidth int = w
	var rectHeight int = h
	var terrain [][]int = make([][]int, h)

	for rectY := y; rectY < y+rectHeight; rectY++ {
		terrain[rectY] = make([]int, 0)

		for rectX := x; rectX < x+rectWidth; rectX++ {
			terrain[rectY] = append(terrain[rectY], plane.Map.Layers[z].Terrain[rectY][rectX])
		}
	}

	return terrain
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

			if zone.Id == zoneId {
				plane.Zone = zone
			}
		}

		if zoneId != 0 && plane.Zone == nil {
			return errors.New("trying to load plane with a bad zone")
		}

		game.Planes.Insert(plane)
	}

	log.Printf("Loaded %d planes from database.\r\n", game.Planes.Count)
	return nil
}
