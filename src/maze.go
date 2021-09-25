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
	"math/rand"
	"strings"
)

type MazeCell struct {
	grid *MazeGrid
	room *Room
	wall bool
	x    int
	y    int
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
				room: nil,
				grid: maze,
				x:    x,
				y:    y,
				wall: true,
			}
		}
	}

	return maze
}

func (grid *MazeGrid) isValidPosition(x int, y int) bool {
	return x >= 1 && x < grid.width-1 && y >= 1 && y < grid.height-1
}

func (cell *MazeCell) getAdjacentCells(wall bool, distance int) *LinkedList {
	list := NewLinkedList()

	for direction := DirectionNorth; direction < DirectionUp; direction++ {
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
		}

		if cell.grid.isValidPosition(translatedX, translatedY) && cell.grid.grid[translatedX][translatedY].wall == wall {
			list.Insert(cell.grid.grid[translatedX][translatedY])
		}
	}

	return list
}

/* Dig a maze using Prim's algorithm */
func (maze *MazeGrid) generatePrimMaze() {
	maze.entryX = rand.Intn(maze.width - 1)
	maze.entryY = rand.Intn(maze.height - 1)

	var entryPoint *MazeCell = maze.grid[maze.entryX][maze.entryY]

	entryPoint.setWall(false)
	frontiers := entryPoint.getAdjacentCells(true, 2)

	for {
		if len(frontiers.Values()) < 1 {
			break
		}

		f := frontiers.GetRandomNode().Value.(*MazeCell)
		neighbours := f.getAdjacentCells(false, 2)

		if neighbours.Count > 0 {
			neighbour := neighbours.GetRandomNode().Value.(*MazeCell)

			passageX := (neighbour.x + f.x) / 2
			passageY := (neighbour.y + f.y) / 2

			f.setWall(false)
			maze.grid[passageX][passageY].setWall(false)
			neighbour.setWall(false)
		}

		frontierCells := f.getAdjacentCells(true, 2)
		frontiers.Concat(frontierCells)
		frontiers.Remove(f)

		continue
	}
}

func (cell *MazeCell) setWall(wall bool) {
	cell.wall = false
}

func (game *Game) doMazeTesting() {
	/* Generate a five-floor dungeon every runtime to test the algorithms */
	dungeon := game.GenerateDungeon(5)

	/* Hardcode an exit from limbo into the first floor of the test dungeon */
	limbo, err := game.LoadRoomIndex(RoomLimbo)
	if err != nil {
		log.Println(err)
	}

	if len(dungeon.floors) < 1 {
		log.Printf("Failed to generate dungeons, cannot link to limbo; aborting maze testing.\r\n")
		return
	}

	maze := dungeon.floors[0]
	maze.print()

	limbo.exit[DirectionDown] = &Exit{
		id:        0,
		direction: DirectionDown,
		to:        maze.grid[maze.entryX][maze.entryY].room,
		flags:     EXIT_IS_DOOR | EXIT_CLOSED,
	}
	maze.grid[maze.entryX][maze.entryY].room.exit[DirectionUp] = &Exit{
		id:        0,
		direction: DirectionUp,
		to:        limbo,
		flags:     EXIT_IS_DOOR | EXIT_CLOSED,
	}
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

func (maze *MazeGrid) print() {
	var output strings.Builder

	for y := 0; y < maze.height; y++ {
		for x := 0; x < maze.width; x++ {
			if maze.grid[x][y].wall {
				output.WriteString("#")
			} else {
				if x == maze.entryX && y == maze.entryY {
					output.WriteString("*")
				} else if x == maze.endX && y == maze.endY {
					output.WriteString("X")
				} else {
					output.WriteString(" ")
				}
			}
		}

		output.WriteString("\r\n")
	}

	fmt.Println(output.String())
}
