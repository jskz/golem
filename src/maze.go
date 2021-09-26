/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"log"
	"math/rand"
	"strings"
	"sync"
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
	maze.entryX = rand.Intn(maze.width-2) + 1
	maze.entryY = rand.Intn(maze.height-2) + 1

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
	/* Hardcode an exit from limbo into the first floor of the test dungeon */
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
			/* Generate a five-floor dungeon every runtime to test the algorithms */
			dungeon = game.GenerateDungeon(5, 30, 30)

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
			if maze.grid[x][y].wall {
				output.WriteString("{D#")
			} else {
				if x == ch.Room.cell.x && y == ch.Room.cell.y {
					output.WriteString("{Y@")
				} else if x == maze.entryX && y == maze.entryY {
					output.WriteString("{M^")
				} else if x == maze.endX && y == maze.endY {
					output.WriteString("{Mv")
				} else {
					output.WriteString("{c.")
				}
			}
		}

		output.WriteString("\r\n")
	}

	return output.String()
}
