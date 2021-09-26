/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "log"

/*
 * A dungeon represents a related set of mazes connected to one another in sequence.
 *
 * This abstraction will make available to the maze generator information about the
 * floor, overarching theme, etc.
 */
type Dungeon struct {
	game     *Game       `json:"game"`
	floors   []*MazeGrid `json:"floors"`
	entrance *Room       `json:"entrance"`
	abyss    *Room       `json:"abyss"`
}

func (game *Game) GenerateDungeon(floorCount int, dungeonWidth int, dungeonHeight int) *Dungeon {
	dungeon := &Dungeon{game: game}
	dungeon.floors = make([]*MazeGrid, 0)

	if floorCount < 1 {
		return nil
	}

	var previousFloorExit *Room = nil
	log.Printf("Generating a %d floor dungeon of dimensions %dx%d\r\n", floorCount, dungeonWidth, dungeonHeight)

	for i := 0; i < floorCount; i++ {
		floor := game.NewMaze(dungeonWidth, dungeonHeight)
		floor.generatePrimMaze()
		floor.reify() /* Ensure the floor's rooms exist before we start populating them */

		if previousFloorExit != nil {
			/* Dig a two-way closed door exit between this room and the "end" of the previous floor */
			previousFloorExit.exit[DirectionDown] = &Exit{
				id:        0,
				direction: DirectionDown,
				to:        floor.grid[floor.entryX][floor.entryY].room,
				flags:     EXIT_IS_DOOR | EXIT_CLOSED,
			}

			floor.grid[floor.entryX][floor.entryY].room.exit[DirectionUp] = &Exit{
				id:        0,
				direction: DirectionUp,
				to:        previousFloorExit,
				flags:     EXIT_IS_DOOR | EXIT_CLOSED,
			}
		}

		/* Find a sufficiently expensive path from the floor's entry point to the entry point of the next floor */
		floor.endX = floor.entryX
		floor.endY = floor.entryY

		entryPoint := floor.grid[floor.entryX][floor.entryY]
		fScore := 0

		for y := 0; y < floor.height; y++ {
			for x := 0; x < floor.width; x++ {
				if !floor.grid[x][y].wall && floor.grid[x][y].room != nil {
					nodes := floor.findPathAStar(entryPoint, floor.grid[x][y])
					difficulty := len(nodes) - 1

					if difficulty < 0 {
						difficulty = 0
					}

					if difficulty > fScore {
						fScore = difficulty

						floor.endX = x
						floor.endY = y

						previousFloorExit = floor.grid[x][y].room

						/* The "abyss" is the deepest room in the dungeon's deepest floor */
						dungeon.abyss = previousFloorExit
					}
				}
			}
		}

		if floor.endX == floor.entryX && floor.endY == floor.entryY {
			log.Printf("Could not find a suitable maze path while generating a dungeon, aborting.\r\n")
			break
		}

		log.Printf("Finished generating floor %d, start at (%d, %d) end at (%d, %d): %d difficulty.\r\n", i+1, floor.entryX, floor.entryY, floor.endX, floor.endY, fScore)
		dungeon.floors = append(dungeon.floors, floor)
	}

	dungeon.entrance = dungeon.floors[0].grid[dungeon.floors[0].entryX][dungeon.floors[0].entryY].room
	return dungeon
}
