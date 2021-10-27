/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"math/rand"
	"strings"
)

type MazeCell struct {
	Grid    *MazeGrid `json:"grid"`
	Room    *Room     `json:"room"`
	Terrain int       `json:"terrain"`
	Wall    bool      `json:"wall"`
	X       int       `json:"x"`
	Y       int       `json:"y"`
}

type MazeGrid struct {
	Game   *Game         `json:"game"`
	Grid   [][]*MazeCell `json:"grid"`
	Width  int           `json:"width"`
	Height int           `json:"height"`
	EntryX int           `json:"entryX"`
	EntryY int           `json:"entryY"`
	EndX   int           `json:"endX"`
	EndY   int           `json:"endY"`
}

func (game *Game) NewMaze(width int, height int) *MazeGrid {
	maze := &MazeGrid{
		Game:   game,
		Grid:   make([][]*MazeCell, width),
		Width:  width,
		Height: height,
	}

	for i := 0; i < height; i++ {
		maze.Grid[i] = make([]*MazeCell, height)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			maze.Grid[x][y] = &MazeCell{
				Room:    nil,
				Grid:    maze,
				X:       x,
				Y:       y,
				Wall:    true,
				Terrain: TerrainTypeCaveDeepWall1,
			}

			if rand.Intn(10) > 7 && x >= 1 && y >= 1 && x <= maze.Width-1 && y <= maze.Height-1 {
				maze.Grid[x][y].Terrain = TerrainTypeCaveDeepWall1 + rand.Intn(4)
			}
		}
	}

	return maze
}

func (grid *MazeGrid) isValidPosition(x int, y int) bool {
	return x >= 1 && x < grid.Width-1 && y >= 1 && y < grid.Height-1
}

func (cell *MazeCell) getAdjacentCells(wall bool, distance int, ordinals bool) *LinkedList {
	list := NewLinkedList()

	top := DirectionMax
	if !ordinals {
		top = DirectionNortheast
	}

	for direction := DirectionNorth; direction < top; direction++ {
		var translatedX int = cell.X
		var translatedY int = cell.Y

		switch direction {
		case DirectionNorth:
			translatedX = cell.X
			translatedY = cell.Y - distance
		case DirectionEast:
			translatedX = cell.X + distance
			translatedY = cell.Y
		case DirectionSouth:
			translatedX = cell.X
			translatedY = cell.Y + distance
		case DirectionWest:
			translatedX = cell.X - distance
			translatedY = cell.Y
		case DirectionNortheast:
			translatedX = cell.X + distance
			translatedY = cell.Y - distance
		case DirectionSoutheast:
			translatedX = cell.X + distance
			translatedY = cell.Y + distance
		case DirectionSouthwest:
			translatedX = cell.X - distance
			translatedY = cell.Y + distance
		case DirectionNorthwest:
			translatedX = cell.X - distance
			translatedY = cell.Y - distance
		}

		if cell.Grid.isValidPosition(translatedX, translatedY) && cell.Grid.Grid[translatedX][translatedY].Wall == wall {
			list.Insert(cell.Grid.Grid[translatedX][translatedY])
		}
	}

	return list
}

func (cell *MazeCell) setAdjacentCellsTerrainType(wall bool, distance int, terrain int) {
	cells := cell.getAdjacentCells(wall, distance, true)

	for iter := cells.Head; iter != nil; iter = iter.Next {
		cell := iter.Value.(*MazeCell)

		if cell.Wall == wall {
			cell.Terrain = terrain
		}
	}
}

/* Dig a maze using Prim's algorithm */
func (maze *MazeGrid) generatePrimMaze() {
	maze.EntryX = rand.Intn(maze.Width-3) + 2
	maze.EntryY = rand.Intn(maze.Height-3) + 2

	var entryPoint *MazeCell = maze.Grid[maze.EntryX][maze.EntryY]

	entryPoint.setWall(false)
	entryPoint.Terrain = TerrainTypeCaveTunnel

	frontiers := entryPoint.getAdjacentCells(true, 2, false)

	for {
		if len(frontiers.Values()) < 1 {
			break
		}

		f := frontiers.GetRandomNode().Value.(*MazeCell)
		neighbours := f.getAdjacentCells(false, 2, false)

		if neighbours.Count > 0 {
			neighbour := neighbours.GetRandomNode().Value.(*MazeCell)

			passageX := (neighbour.X + f.X) / 2
			passageY := (neighbour.Y + f.Y) / 2

			f.setWall(false)
			f.Terrain = TerrainTypeCaveTunnel
			maze.Grid[passageX][passageY].setWall(false)
			maze.Grid[passageX][passageY].Terrain = TerrainTypeCaveTunnel
			neighbour.setWall(false)
			neighbour.Terrain = TerrainTypeCaveTunnel

			f.setAdjacentCellsTerrainType(true, 1, TerrainTypeCaveWall)
			maze.Grid[passageX][passageY].setAdjacentCellsTerrainType(true, 1, TerrainTypeCaveWall)
			neighbour.setAdjacentCellsTerrainType(true, 1, TerrainTypeCaveWall)
		}

		frontierCells := f.getAdjacentCells(true, 2, false)
		frontiers.Concat(frontierCells)
		frontiers.Remove(f)

		continue
	}
}

func (cell *MazeCell) setWall(wall bool) {
	cell.Wall = wall
}

func (maze *MazeGrid) createRoom(x int, y int) *Room {
	if maze.Grid[x][y].Room != nil {
		return maze.Grid[x][y].Room
	}

	room := maze.Game.NewRoom()
	room.Id = 0
	room.Zone = nil
	room.Virtual = true
	room.Cell = maze.Grid[x][y]
	room.Name = "In the Underground"
	room.Description = "You are deep within the dark dungeons of development."
	room.Exit = make(map[uint]*Exit)
	room.Characters = NewLinkedList()
	room.Objects = NewLinkedList()

	maze.Grid[x][y].Room = room
	return room
}

func (maze *MazeGrid) reify() {
	for y := 0; y < maze.Height; y++ {
		for x := 0; x < maze.Width; x++ {
			if !maze.Grid[x][y].Wall {
				room := maze.createRoom(x, y)

				for direction := DirectionNorth; direction < DirectionUp; direction++ {
					var translatedX int = x
					var translatedY int = y

					switch direction {
					case DirectionNorth:
						translatedX = x
						translatedY = y - 1
					case DirectionEast:
						translatedX = x + 1
						translatedY = y
					case DirectionSouth:
						translatedX = x
						translatedY = y + 1
					case DirectionWest:
						translatedX = x - 1
						translatedY = y
					}

					if maze.isValidPosition(translatedX, translatedY) && !maze.Grid[translatedX][translatedY].Wall {
						to := maze.createRoom(translatedX, translatedY)

						exit := &Exit{}
						exit.To = to
						exit.Direction = uint(direction)

						room.Exit[exit.Direction] = exit
					}
				}
			}
		}
	}
}

func (ch *Character) CreateMazeMap() string {
	var output strings.Builder

	if ch.Room == nil || !ch.Room.Virtual || ch.Room.Cell == nil {
		return ""
	}

	var maze *MazeGrid = ch.Room.Cell.Grid
	if maze == nil {
		return ""
	}

	for y := 0; y < maze.Height; y++ {
		for x := 0; x < maze.Width; x++ {
			if x == ch.Room.Cell.X && y == ch.Room.Cell.Y {
				output.WriteString("{Y@")
			} else if x == maze.EntryX && y == maze.EntryY {
				output.WriteString("{M^")
			} else if x == maze.EndX && y == maze.EndY {
				output.WriteString("{Mv")
			} else {
				var terrain *Terrain = TerrainTable[maze.Grid[x][y].Terrain]

				output.WriteString(terrain.MapGlyph)
			}
		}

		output.WriteString("\r\n")
	}

	return output.String()
}
