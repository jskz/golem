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
	grid    *MazeGrid
	room    *Room
	terrain int
	wall    bool
	x       int
	y       int
}

type MazeGrid struct {
	game   *Game
	grid   [][]*MazeCell
	width  int
	height int
	entryX int
	entryY int
	endX   int
	endY   int
}

func (game *Game) NewMaze(width int, height int) *MazeGrid {
	maze := &MazeGrid{
		game:   game,
		grid:   make([][]*MazeCell, width),
		width:  width,
		height: height,
	}

	for i := 0; i < height; i++ {
		maze.grid[i] = make([]*MazeCell, height)
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			maze.grid[x][y] = &MazeCell{
				room:    nil,
				grid:    maze,
				x:       x,
				y:       y,
				wall:    true,
				terrain: TerrainTypeCaveDeepWall1,
			}

			if rand.Intn(10) > 7 && x >= 1 && y >= 1 && x <= maze.width-1 && y <= maze.height-1 {
				maze.grid[x][y].terrain = TerrainTypeCaveDeepWall1 + rand.Intn(4)
			}
		}
	}

	return maze
}

func (grid *MazeGrid) isValidPosition(x int, y int) bool {
	return x >= 1 && x < grid.width-1 && y >= 1 && y < grid.height-1
}

func (cell *MazeCell) getAdjacentCells(wall bool, distance int, ordinals bool) *LinkedList {
	list := NewLinkedList()

	top := DirectionMax
	if !ordinals {
		top = DirectionNortheast
	}

	for direction := DirectionNorth; direction < top; direction++ {
		var translatedX int = cell.x
		var translatedY int = cell.y

		switch direction {
		case DirectionNorth:
			translatedX = cell.x
			translatedY = cell.y - distance
		case DirectionEast:
			translatedX = cell.x + distance
			translatedY = cell.y
		case DirectionSouth:
			translatedX = cell.x
			translatedY = cell.y + distance
		case DirectionWest:
			translatedX = cell.x - distance
			translatedY = cell.y
		case DirectionNortheast:
			translatedX = cell.x + distance
			translatedY = cell.y - distance
		case DirectionSoutheast:
			translatedX = cell.x + distance
			translatedY = cell.y + distance
		case DirectionSouthwest:
			translatedX = cell.x - distance
			translatedY = cell.y + distance
		case DirectionNorthwest:
			translatedX = cell.x - distance
			translatedY = cell.y - distance
		}

		if cell.grid.isValidPosition(translatedX, translatedY) && cell.grid.grid[translatedX][translatedY].wall == wall {
			list.Insert(cell.grid.grid[translatedX][translatedY])
		}
	}

	return list
}

func (cell *MazeCell) setAdjacentCellsTerrainType(wall bool, distance int, terrain int) {
	cells := cell.getAdjacentCells(wall, distance, true)

	for iter := cells.Head; iter != nil; iter = iter.Next {
		cell := iter.Value.(*MazeCell)

		if cell.wall == wall {
			cell.terrain = terrain
		}
	}
}

/* Dig a maze using Prim's algorithm */
func (maze *MazeGrid) generatePrimMaze() {
	maze.entryX = rand.Intn(maze.width-3) + 2
	maze.entryY = rand.Intn(maze.height-3) + 2

	var entryPoint *MazeCell = maze.grid[maze.entryX][maze.entryY]

	entryPoint.setWall(false)
	entryPoint.terrain = TerrainTypeCaveTunnel

	frontiers := entryPoint.getAdjacentCells(true, 2, false)

	for {
		if len(frontiers.Values()) < 1 {
			break
		}

		f := frontiers.GetRandomNode().Value.(*MazeCell)
		neighbours := f.getAdjacentCells(false, 2, false)

		if neighbours.Count > 0 {
			neighbour := neighbours.GetRandomNode().Value.(*MazeCell)

			passageX := (neighbour.x + f.x) / 2
			passageY := (neighbour.y + f.y) / 2

			f.setWall(false)
			f.terrain = TerrainTypeCaveTunnel
			maze.grid[passageX][passageY].setWall(false)
			maze.grid[passageX][passageY].terrain = TerrainTypeCaveTunnel
			neighbour.setWall(false)
			neighbour.terrain = TerrainTypeCaveTunnel

			f.setAdjacentCellsTerrainType(true, 1, TerrainTypeCaveWall)
			maze.grid[passageX][passageY].setAdjacentCellsTerrainType(true, 1, TerrainTypeCaveWall)
			neighbour.setAdjacentCellsTerrainType(true, 1, TerrainTypeCaveWall)
		}

		frontierCells := f.getAdjacentCells(true, 2, false)
		frontiers.Concat(frontierCells)
		frontiers.Remove(f)

		continue
	}
}

func (cell *MazeCell) setWall(wall bool) {
	cell.wall = wall
}

func (maze *MazeGrid) createRoom(x int, y int) *Room {
	if maze.grid[x][y].room != nil {
		return maze.grid[x][y].room
	}

	room := maze.game.NewRoom()
	room.Id = 0
	room.zone = nil
	room.virtual = true
	room.cell = maze.grid[x][y]
	room.name = "In the Underground"
	room.description = "You are deep within the dark dungeons of development."
	room.exit = make(map[uint]*Exit)
	room.Characters = NewLinkedList()
	room.objects = NewLinkedList()

	maze.grid[x][y].room = room
	return room
}

func (maze *MazeGrid) reify() {
	for y := 0; y < maze.height; y++ {
		for x := 0; x < maze.width; x++ {
			if !maze.grid[x][y].wall {
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

					if maze.isValidPosition(translatedX, translatedY) && !maze.grid[translatedX][translatedY].wall {
						to := maze.createRoom(translatedX, translatedY)

						exit := &Exit{}
						exit.to = to
						exit.direction = uint(direction)

						room.exit[exit.direction] = exit
					}
				}
			}
		}
	}
}

func (ch *Character) CreateMazeMap() string {
	var output strings.Builder

	if ch.Room == nil || !ch.Room.virtual || ch.Room.cell == nil {
		return ""
	}

	var maze *MazeGrid = ch.Room.cell.grid
	if maze == nil {
		return ""
	}

	for y := 0; y < maze.height; y++ {
		for x := 0; x < maze.width; x++ {
			if x == ch.Room.cell.x && y == ch.Room.cell.y {
				output.WriteString("{Y@")
			} else if x == maze.entryX && y == maze.entryY {
				output.WriteString("{M^")
			} else if x == maze.endX && y == maze.endY {
				output.WriteString("{Mv")
			} else {
				var terrain *Terrain = TerrainTable[maze.grid[x][y].terrain]

				output.WriteString(terrain.mapGlyph)
			}
		}

		output.WriteString("\r\n")
	}

	return output.String()
}
