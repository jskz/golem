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
	game   *Game
	floors []*MazeGrid
}

func (game *Game) GenerateDungeon(floorCount int) *Dungeon {
	dungeon := &Dungeon{game: game}
	dungeon.floors = make([]*MazeGrid, 0)

	/* Do not have to be constants */
	const dungeonWidth = 30
	const dungeonHeight = 30

	for i := 0; i < floorCount; i++ {
		floor := game.NewMaze(dungeonWidth, dungeonHeight)

		dungeon.floors = append(dungeon.floors, floor)
	}

	return dungeon
}
