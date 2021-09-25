/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

/*
 * A dungeon represents a related set of mazes connected to one another in sequence.
 *
 * This abstraction will make available to the maze generator information about the
 * floor, overarching theme, etc.
 */
type Dungeon struct {
	game     *Game
	floors   []*MazeGrid
	entrance *Room
	abyss    *Room
}

func (game *Game) GenerateDungeon(floorCount int) *Dungeon {
	dungeon := &Dungeon{game: game}
	dungeon.floors = make([]*MazeGrid, 0)

	/* Do not have to be constants */
	const dungeonWidth = 30
	const dungeonHeight = 30

	var previousFloorExit *Room = nil

	for i := 0; i < floorCount; i++ {
		floor := game.NewMaze(dungeonWidth, dungeonHeight)
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
				direction: DirectionDown,
				to:        floor.grid[floor.entryX][floor.entryY].room,
				flags:     EXIT_IS_DOOR | EXIT_CLOSED,
			}
		}

		/* Find a sufficiently expensive path from the floor's entry point to the entry point of the next floor */

		/* previousFloorExit = ... */

		dungeon.floors = append(dungeon.floors, floor)
	}

	return dungeon
}
