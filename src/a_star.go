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
	"math"
	"strconv"
	"strings"
)

type MazeAStarVisit struct {
	gScore   int
	fScore   int
	previous *MazeAStarVisit
	cell     *MazeCell
}

/* Taxi cab distance because only cardinal directions for now :-) */
func (maze *MazeGrid) heuristic(node *MazeCell, target *MazeCell) int {
	if node == nil || target == nil {
		return 0
	}

	d := math.Abs(float64(node.Y-target.Y)) + math.Abs(float64(node.X-target.X))

	return int(d)
}

func do_path(ch *Character, arguments string) {
	if ch.Room == nil || !ch.Room.Virtual || ch.Room.Cell == nil || ch.Room.Cell.Grid == nil {
		ch.Send("You cannot pathfind from here.\r\n")
		return
	}

	args := strings.Split(arguments, " ")
	if len(args) < 2 {
		ch.Send(fmt.Sprintf("Usage: path <x> <y>\r\nYour current position is (%d, %d).\r\n", ch.Room.Cell.X, ch.Room.Cell.Y))
		return
	}

	x, err := strconv.Atoi(args[0])
	if err != nil {
		ch.Send(fmt.Sprintf("Usage: path <x> <y>\r\nYour current position is (%d, %d).\r\n", ch.Room.Cell.X, ch.Room.Cell.Y))
		return
	}

	y, err := strconv.Atoi(args[1])
	if err != nil {
		ch.Send(fmt.Sprintf("Usage: path <x> <y>\r\nYour current position is (%d, %d).\r\n", ch.Room.Cell.X, ch.Room.Cell.Y))
		return
	}

	grid := ch.Room.Cell.Grid
	if !grid.isValidPosition(x, y) {
		ch.Send(fmt.Sprintf("Target (%d, %d) out of bounds.\r\n", x, y))
		return
	}

	target := grid.cellAt(x, y)
	if target == nil || target.Wall {
		ch.Send(fmt.Sprintf("Bad cell or obstacle at (%d, %d).\r\n", x, y))
		return
	}

	var output strings.Builder

	pathNodes := grid.findPathAStar(ch.Room.Cell, target)
	if len(pathNodes) == 0 && ch.Room.Cell != target {
		ch.Send(fmt.Sprintf("No path from (%d, %d) to (%d, %d).\r\n", ch.Room.Cell.X, ch.Room.Cell.Y, target.X, target.Y))
		return
	}

	output.WriteString(fmt.Sprintf("{YPath from (%d, %d) to (%d, %d) in %d moves.{x\r\n", ch.Room.Cell.X, ch.Room.Cell.Y, target.X, target.Y, int(math.Max(0, float64(len(pathNodes)-1)))))
	for r := len(pathNodes) - 1; r >= 0; r-- {
		output.WriteString(fmt.Sprintf("{G(%d, %d){x\r\n", pathNodes[r].cell.X, pathNodes[r].cell.Y))
	}

	ch.Send(output.String())
}

func (maze *MazeGrid) findPathAStar(start *MazeCell, end *MazeCell) []*MazeAStarVisit {
	visits := make([]*MazeAStarVisit, 0)
	if maze == nil || start == nil || end == nil || start.Grid != maze || end.Grid != maze || start.Wall || end.Wall {
		return visits
	}

	if start == end {
		return visits
	}

	visited := make(map[*MazeCell]bool)
	unvisited := make(map[*MazeCell]*MazeAStarVisit)

	for y := 0; y < maze.Height; y++ {
		for x := 0; x < maze.Width; x++ {
			cell := maze.cellAt(x, y)
			if cell == nil || cell.Wall {
				continue
			}

			unvisited[cell] = &MazeAStarVisit{gScore: 1000000, fScore: 1000000, previous: nil, cell: cell}
		}
	}

	startVisit, ok := unvisited[start]
	if !ok {
		return visits
	}

	if _, ok := unvisited[end]; !ok {
		return visits
	}

	startVisit.gScore = 0
	startVisit.fScore = maze.heuristic(start, end)
	startVisit.previous = nil

	for {
		if len(unvisited) == 0 {
			return visits
		} else {
			var currentNode *MazeAStarVisit = nil

			for _, visit := range unvisited {
				if currentNode == nil || visit.fScore < currentNode.fScore {
					currentNode = visit
				}
			}

			if currentNode.cell == end {
				for current := currentNode; current != nil; current = current.previous {
					visits = append(visits, current)
					if current.cell == start {
						return visits
					}
				}

				return make([]*MazeAStarVisit, 0)
			} else {
				neighbours := currentNode.cell.getAdjacentCells(false, 1, false)
				for neighbour := range neighbours.All() {
					if !visited[neighbour] {
						var gScore int = currentNode.gScore + 1

						neighbourVisit, ok := unvisited[neighbour]
						if ok {
							if gScore < neighbourVisit.gScore {
								neighbourVisit.gScore = gScore
								neighbourVisit.fScore = gScore + maze.heuristic(neighbour, end)
								neighbourVisit.previous = currentNode
							}
						}
					}
				}

				visited[currentNode.cell] = true
				delete(unvisited, currentNode.cell)
			}
		}
	}
}
