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
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
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

type District struct {
	Layer              *MapGrid       `json:"layer"`
	Id                 int            `json:"id"`
	Plane              *Plane         `json:"plane"`
	Rect               *Rect          `json:"rect"`
	TerrainNameMapping map[int]string `json:"terrainNameMapping"`
}

type PlaneObserver struct {
	Plane *Plane `json:"plane"`
	Rect  *Rect  `json:"rect"`

	OnEnterCallback goja.Callable `json:"onEnterCallback"`
	OnLeaveCallback goja.Callable `json:"onLeaveCallback"`
}

type MapGrid struct {
	Observers []*PlaneObserver `json:"observers"`
	Districts *LinkedList      `json:"districts"`

	Terrain [][]int `json:"terrain"`
	Atlas   *Atlas  `json:"atlas"`
	Width   int     `json:"width"`
	Height  int     `json:"height"`
}

type Map struct {
	Layers []*MapGrid `json:"layers"`
}

// The "Atlas" structure is:
//
//   - A collection of maps for a Plane between (x, y) points expressed as an integer
//     and linked lists for game objects: characters, objects, rooms, exits.
//   - A collection of related quadtrees allowing an interface for easy spatial queries on
//     the same data.
type Atlas struct {
	Plane *Plane `json:"plane"`

	// TODO: portals, scripts
	Characters map[int]*LinkedList    `json:"characters"`
	Objects    map[int]*LinkedList    `json:"objects"`
	Rooms      map[int]*LinkedList    `json:"rooms"`
	Exits      map[int]map[uint]*Exit `json:"exits"`

	CharacterTree *QuadTree `json:"characterTree"`
	ObjectTree    *QuadTree `json:"objectTree"`
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

func (plane *Plane) SupportsPersistentCoordinates() bool {
	if plane == nil {
		return false
	}

	return plane.SourceType == SourceTypeBlob
}

func (plane *Plane) NewAtlas() *Atlas {
	return &Atlas{
		Plane:      plane,
		Characters: make(map[int]*LinkedList),
		Objects:    make(map[int]*LinkedList),
		Exits:      make(map[int]map[uint]*Exit),

		CharacterTree: NewQuadTree(float64(plane.Width), float64(plane.Height)),
		ObjectTree:    NewQuadTree(float64(plane.Width), float64(plane.Height)),
	}
}

func (plane *Plane) terrainBlobSize() int {
	return plane.Depth * plane.Height * plane.Width
}

func (plane *Plane) terrainBlobIndex(x int, y int, z int) int {
	return z*plane.Height*plane.Width + y*plane.Width + x
}

func (plane *Plane) atlasKey(x int, y int) int {
	return y*plane.Width + x
}

func (plane *Plane) containsCoordinates(x int, y int, z int) bool {
	return x >= 0 && x < plane.Width &&
		y >= 0 && y < plane.Height &&
		z >= 0 && z < plane.Depth
}

func (plane *Plane) newMapGrid() *MapGrid {
	grid := &MapGrid{Atlas: plane.NewAtlas(), Districts: NewLinkedList()}
	grid.Terrain = make([][]int, plane.Height)

	for y := 0; y < plane.Height; y++ {
		grid.Terrain[y] = make([]int, plane.Width)
	}

	return grid
}

func (plane *Plane) terrainMapFromBlob(blob []byte) *Map {
	planeMap := &Map{
		Layers: make([]*MapGrid, 0, plane.Depth),
	}

	for z := 0; z < plane.Depth; z++ {
		grid := plane.newMapGrid()

		for y := 0; y < plane.Height; y++ {
			for x := 0; x < plane.Width; x++ {
				grid.Terrain[y][x] = int(blob[plane.terrainBlobIndex(x, y, z)])
			}
		}

		planeMap.Layers = append(planeMap.Layers, grid)
	}

	return planeMap
}

func (plane *Plane) initializeTerrainBlob(defaultTerrain int) ([]byte, int) {
	plane.Map = &Map{
		Layers: make([]*MapGrid, 0, plane.Depth),
	}

	blob := make([]byte, plane.terrainBlobSize())

	for z := 0; z < plane.Depth; z++ {
		grid := plane.newMapGrid()

		for y := 0; y < plane.Height; y++ {
			for x := 0; x < plane.Width; x++ {
				grid.Terrain[y][x] = defaultTerrain
				blob[plane.terrainBlobIndex(x, y, z)] = byte(defaultTerrain)
			}
		}

		plane.Map.Layers = append(plane.Map.Layers, grid)
	}

	return blob, len(blob)
}

func (plane *Plane) terrainBlobFromMap() []byte {
	blob := make([]byte, plane.terrainBlobSize())

	for z := 0; z < plane.Depth; z++ {
		for y := 0; y < plane.Height; y++ {
			for x := 0; x < plane.Width; x++ {
				blob[plane.terrainBlobIndex(x, y, z)] = byte(plane.Map.Layers[z].Terrain[y][x])
			}
		}
	}

	return blob
}

func (obs *PlaneObserver) Dispose() {
}

func (layer *MapGrid) FindDistrict(x int, y int) *District {
	for iter := layer.Districts.Head; iter != nil; iter = iter.Next {
		d := iter.Value.(*District)

		if d.Rect.Contains(float64(x), float64(y)) {
			return d
		}
	}

	return nil
}

func (plane *Plane) districtLayer(z int) (*MapGrid, bool) {
	if plane == nil || plane.Map == nil || z < 0 || z >= len(plane.Map.Layers) {
		return nil, false
	}

	layer := plane.Map.Layers[z]
	if layer == nil {
		return nil, false
	}

	if layer.Districts == nil {
		layer.Districts = NewLinkedList()
	}

	return layer, true
}

func (layer *MapGrid) RegisterObserver(rect *Rect, options goja.Object, onEnterCallback goja.Callable, onLeaveCallback goja.Callable) goja.Value {
	obs := &PlaneObserver{Plane: layer.Atlas.Plane, Rect: rect, OnEnterCallback: onEnterCallback, OnLeaveCallback: onLeaveCallback}
	layer.Observers = append(layer.Observers, obs)

	game := layer.Atlas.Plane.Game
	observerHandle := game.vm.NewObject()

	observerHandle.Set("dispose", game.vm.ToValue(func() goja.Value {
		for i, observer := range layer.Observers {
			if observer == obs {
				layer.Observers = append(layer.Observers[:i], layer.Observers[i+1:]...)
				return game.vm.ToValue(true)
			}
		}

		return game.vm.ToValue(false)
	}))

	// TODO: Create a stateful "observer" object and pass back in retval:
	//
	// - a setRect() method to post-hoc move/resize the observer "camera"

	return game.vm.ToValue(observerHandle)
}

func (plane *Plane) SaveBlob() error {
	log.Printf("Saving blob for plane %d.\r\n", plane.Id)

	blob := plane.terrainBlobFromMap()

	_, err := plane.Game.db.Exec(`
		UPDATE
			planes
		SET
			source_value = ?
		WHERE
			id = ?
	`, blob, plane.Id)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Saved blob %d.\r\n", plane.Id)
	return nil
}

// Fill the source_value field for this plane with an appropriately sized binary blob.
func (plane *Plane) InitializeBlob() ([]byte, int, error) {
	log.Printf("Initializing new blob for plane %d.\r\n", plane.Id)

	blob, blobSize := plane.initializeTerrainBlob(TerrainTypeOcean)

	_, err := plane.Game.db.Exec(`
		UPDATE
			planes
		SET
			source_value = ?
		WHERE
			id = ?
	`, blob, plane.Id)
	if err != nil {
		return nil, 0, err
	}

	return blob, blobSize, nil
}

func (ch *Character) CreatePlaneMap() string {
	if ch.Room == nil || ch.Room.Plane == nil {
		return "Error retrieving plane map\r\n"
	}

	var buf strings.Builder

	var cameraWidth int = 48
	var cameraHeight int = 18
	var cameraRange int = 9

	cameraX := ch.Room.X
	cameraY := ch.Room.Y
	cameraZ := ch.Room.Z
	lastColour := ""

	for cY := cameraY - (cameraHeight / 2); cY < cameraY+(cameraHeight/2)+1; cY++ {
		for cX := cameraX - (cameraWidth / 2); cX < cameraX+(cameraWidth/2); cX++ {
			if cX < 0 || cX >= ch.Room.Plane.Width || cY < 0 || cY >= ch.Room.Plane.Height || Distance2D(float64(cameraX), float64(cameraY), float64(cX), float64(cY), 2.4, 1) > cameraRange {
				buf.WriteString(" ")
				lastColour = " "
				continue
			}

			if ch.Room.X == cX && cY == ch.Room.Y {
				buf.WriteString("{Y@")
				lastColour = "{Y"
				continue
			}

			otherCharacters, ok := ch.Room.Plane.Map.Layers[ch.Room.Z].Atlas.Characters[ch.Room.Plane.atlasKey(cX, cY)]
			if ok {
				if otherCharacters.Count > 0 {
					buf.WriteString("{W@")
					lastColour = "{W"
					continue
				}
			}

			if lastColour == "" || lastColour != TerrainTable[ch.Room.Plane.Map.Layers[cameraZ].Terrain[cY][cX]].GlyphColour {
				lastColour = TerrainTable[ch.Room.Plane.Map.Layers[cameraZ].Terrain[cY][cX]].GlyphColour
				buf.WriteString(lastColour)
			}

			buf.WriteString(TerrainTable[ch.Room.Plane.Map.Layers[cameraZ].Terrain[cY][cX]].MapGlyph)
		}

		buf.WriteString("\r\n")
	}

	return buf.String()
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

			planeMap := plane.terrainMapFromBlob(blob)

			log.Printf("Plane %d (%d,%d) initialized from %d byte blob.\r\n", plane.Id, plane.Width, plane.Height, blobSize)

			plane.Flags |= PLANE_INITIALIZED
			plane.Map = planeMap

			go func() {
				// Code smell, but allow the main thread a little time to poll for this event
				<-time.After(1 * time.Second)
				game.planeGenerationCompleted <- plane.Id
			}()
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
					dungeon = game.GenerateDungeon(plane.Depth, plane.Width, plane.Height, false)

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

func terrainRestrictsOrdinalPlanarLinks(terrainType int) bool {
	switch terrainType {
	case TerrainTypeOverworldCityEntrance:
		return true
	}

	return false
}

func (plane *Plane) MaterializeRoom(x int, y int, z int, src bool) *Room {
	if !plane.containsCoordinates(x, y, z) {
		return nil
	}

	if plane.PlaneType == PlaneTypeMaze {
		if plane.Dungeon == nil || z >= len(plane.Dungeon.Floors) || plane.Dungeon.Floors[z] == nil {
			return nil
		}

		floor := plane.Dungeon.Floors[z]
		if x >= len(floor.Grid) || floor.Grid[x] == nil || y >= len(floor.Grid[x]) || floor.Grid[x][y] == nil {
			return nil
		}

		return floor.Grid[x][y].Room
	}

	if plane.Map == nil || z >= len(plane.Map.Layers) || plane.Map.Layers[z] == nil {
		return nil
	}

	room := plane.Game.NewRoom()
	room.Plane = plane
	room.Id = 0
	room.Name = "Holodeck"
	room.Description = "If you are seeing this message, something has gone wrong."
	room.Flags = ROOM_VIRTUAL | ROOM_PLANAR

	var ok bool = false

	atlasKey := plane.atlasKey(x, y)

	room.Characters, ok = plane.Map.Layers[z].Atlas.Characters[atlasKey]
	if !ok {
		list := NewLinkedList()

		plane.Map.Layers[z].Atlas.Characters[atlasKey] = list
		room.Characters = list
	}

	ok = false
	room.Objects, ok = plane.Map.Layers[z].Atlas.Objects[atlasKey]
	if !ok {
		list := NewLinkedList()

		plane.Map.Layers[z].Atlas.Objects[atlasKey] = list
		room.Objects = list
	}

	ok = false
	room.Exit, ok = plane.Map.Layers[z].Atlas.Exits[atlasKey]
	if !ok {
		exits := make(map[uint]*Exit, DirectionMax)

		plane.Map.Layers[z].Atlas.Exits[atlasKey] = exits
		room.Exit = exits
	}

	room.X = x
	room.Y = y
	room.Z = z

	if src {
		sourceTerrain := plane.Map.Layers[z].Terrain[y][x]

		/* Try to materialize adjacent planar rooms and link them */
		for _, direction := range PlanarDirections {
			delta, ok := PlanarDirectionDeltaFor(direction)
			if !ok {
				continue
			}

			translatedX := x + delta.X
			translatedY := y + delta.Y

			if plane.containsCoordinates(translatedX, translatedY, z) {
				targetTerrain := plane.Map.Layers[z].Terrain[translatedY][translatedX]

				// If this terrain type is impassible, don't try to materialize it
				if TerrainTable[targetTerrain].Flags&TERRAIN_IMPASSABLE != 0 {
					continue
				}

				if DirectionIsOrdinal(direction) &&
					(terrainRestrictsOrdinalPlanarLinks(sourceTerrain) || terrainRestrictsOrdinalPlanarLinks(targetTerrain)) {
					continue
				}
			}

			adj := plane.MaterializeRoom(translatedX, translatedY, z, false)
			if adj == nil {
				continue
			}

			_, ok = room.Exit[uint(direction)]
			if ok {
				continue
			}

			room.Exit[uint(direction)] = &Exit{
				Room:      room,
				Id:        0,
				To:        adj,
				Direction: direction,
				Flags:     0,
			}

			reverseDirection := ReverseDirection[direction]
			adj.Exit[reverseDirection] = &Exit{
				Room:      adj,
				Id:        0,
				To:        room,
				Direction: reverseDirection,
				Flags:     0,
			}
		}
	}

	return room
}

func (plane *Plane) GetTerrainRect(x int, y int, z int, w int, h int) [][]int {
	var terrain [][]int = make([][]int, 0)

	for rectY := y; rectY < y+h; rectY++ {
		row := make([]int, w)

		c := 0
		for rectX := x; rectX < x+w; rectX++ {
			if !plane.containsCoordinates(rectX, rectY, z) {
				row[c] = 0
				c++
				continue
			}

			row[c] = plane.Map.Layers[z].Terrain[rectY][rectX]
			c++
		}

		terrain = append(terrain, row)
	}

	return terrain
}

func (game *Game) FindPlaneByName(name string) *Plane {
	for iter := game.Planes.Head; iter != nil; iter = iter.Next {
		plane := iter.Value.(*Plane)

		if plane.Name == name {
			return plane
		}
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

			if zone.Id == zoneId {
				plane.Zone = zone
			}
		}

		if zoneId != 0 && plane.Zone == nil {
			return errors.New("trying to load plane with a bad zone")
		}

		game.Planes.Insert(plane)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d planes from database.\r\n", game.Planes.Count)
	return nil
}

func (game *Game) LoadDistricts() error {
	log.Printf("Loading districts.\r\n")

	rows, err := game.db.Query(`
		SELECT
			id,
			plane_id,
			x,
			y,
			z,
			width,
			height
		FROM
			districts
	`)
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	count := 0

	for rows.Next() {
		var planeId int
		var z int

		district := &District{
			Rect:               &Rect{},
			TerrainNameMapping: make(map[int]string),
		}

		err := rows.Scan(&district.Id, &planeId, &district.Rect.X, &district.Rect.Y, &z, &district.Rect.W, &district.Rect.H)
		if err != nil {
			return err
		}

		district.Plane = game.FindPlaneByID(planeId)

		if district.Plane == nil {
			log.Printf("Ignoring district with ID %d loaded for plane with bad ID %d.\r\n", district.Id, planeId)
			continue
		}

		layer, ok := district.Plane.districtLayer(z)
		if !ok {
			log.Printf("Ignoring district with ID %d loaded for invalid map layer %d on plane with ID %d.\r\n", district.Id, z, planeId)
			continue
		}

		collides := false

		for iter := layer.Districts.Head; iter != nil; iter = iter.Next {
			d := iter.Value.(*District)

			if d.Rect.CollidesRect(district.Rect) {
				log.Printf("District %d collides with an existing district, ignoring.\r\n", district.Id)
				collides = true
				break
			}
		}

		if collides {
			continue
		}

		district.Layer = layer
		layer.Districts.Insert(district)
		count++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d districts from database.\r\n", count)
	return nil
}

func (game *Game) FindDistrictByID(id int) *District {
	for planeIter := game.Planes.Head; planeIter != nil; planeIter = planeIter.Next {
		plane := planeIter.Value.(*Plane)

		if plane.Map == nil || len(plane.Map.Layers) == 0 {
			continue
		}

		for _, layer := range plane.Map.Layers {
			if layer == nil || layer.Districts == nil {
				continue
			}

			for districtIter := layer.Districts.Head; districtIter != nil; districtIter = districtIter.Next {
				district := districtIter.Value.(*District)

				if district.Id == id {
					return district
				}
			}
		}
	}

	return nil
}
