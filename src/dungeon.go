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
	Game     *Game       `json:"game"`
	Floors   []*MazeGrid `json:"floors"`
	Entrance *Room       `json:"entrance"`
	Abyss    *Room       `json:"abyss"`
}

func (game *Game) GenerateDungeon(floorCount int, dungeonWidth int, dungeonHeight int) *Dungeon {
	dungeon := &Dungeon{Game: game}
	dungeon.Floors = make([]*MazeGrid, 0)

	if floorCount < 1 {
		return nil
	}

	var previousFloorExit *Room = nil
	log.Printf("Generating a %d floor dungeon of dimensions %dx%d\r\n", floorCount, dungeonWidth, dungeonHeight)

	for i := 0; i < floorCount; i++ {
		floor := game.NewMaze(dungeonWidth, dungeonHeight)
		floor.generatePrimMaze()
		floor.reify(i) /* Ensure the floor's rooms exist before we start populating them */

		if previousFloorExit != nil {
			/* Dig a two-way closed door exit between this room and the "end" of the previous floor */
			previousFloorExit.Exit[DirectionDown] = &Exit{
				Id:        0,
				Direction: DirectionDown,
				To:        floor.Grid[floor.EntryX][floor.EntryY].Room,
				Flags:     EXIT_IS_DOOR | EXIT_CLOSED,
			}

			floor.Grid[floor.EntryX][floor.EntryY].Room.Exit[DirectionUp] = &Exit{
				Id:        0,
				Direction: DirectionUp,
				To:        previousFloorExit,
				Flags:     EXIT_IS_DOOR | EXIT_CLOSED,
			}
		}

		/* Find a sufficiently expensive path from the floor's entry point to the entry point of the next floor */
		floor.EndX = floor.EntryX
		floor.EndY = floor.EntryY

		entryPoint := floor.Grid[floor.EntryX][floor.EntryY]
		fScore := 0

		for y := 0; y < floor.Height; y++ {
			for x := 0; x < floor.Width; x++ {
				if !floor.Grid[x][y].Wall && floor.Grid[x][y].Room != nil {
					nodes := floor.findPathAStar(entryPoint, floor.Grid[x][y])
					difficulty := len(nodes) - 1

					if difficulty < 0 {
						difficulty = 0
					}

					if difficulty > fScore {
						fScore = difficulty

						floor.EndX = x
						floor.EndY = y

						previousFloorExit = floor.Grid[x][y].Room

						/* The "abyss" is the deepest room in the dungeon's deepest floor */
						dungeon.Abyss = previousFloorExit
					}
				}
			}
		}

		if floor.EndX == floor.EntryX && floor.EndY == floor.EntryY {
			log.Printf("Could not find a suitable maze path while generating a dungeon, aborting.\r\n")
			break
		}

		log.Printf("Finished generating floor %d, start at (%d, %d) end at (%d, %d): %d difficulty.\r\n", i+1, floor.EntryX, floor.EntryY, floor.EndX, floor.EndY, fScore)
		dungeon.Floors = append(dungeon.Floors, floor)
	}

	dungeon.Entrance = dungeon.Floors[0].Grid[dungeon.Floors[0].EntryX][dungeon.Floors[0].EntryY].Room
	return dungeon
}
