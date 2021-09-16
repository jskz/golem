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
	d := math.Abs(float64(node.y-target.y)) + math.Abs(float64(node.x-target.x))

	return int(d)
}

func do_path(ch *Character, arguments string) {
	if ch.Room == nil || !ch.Room.virtual || ch.Room.cell == nil {
		ch.Send("You cannot pathfind from here.\r\n")
		return
	}

	args := strings.Split(arguments, " ")
	if len(args) < 2 {
		ch.Send(fmt.Sprintf("Usage: path <x> <y>\r\nYour current position is (%d, %d).\r\n", ch.Room.cell.x, ch.Room.cell.y))
		return
	}

	x, err := strconv.Atoi(args[0])
	if err != nil {
		ch.Send(fmt.Sprintf("Usage: path <x> <y>\r\nYour current position is (%d, %d).\r\n", ch.Room.cell.x, ch.Room.cell.y))
		return
	}

	y, err := strconv.Atoi(args[1])
	if err != nil {
		ch.Send(fmt.Sprintf("Usage: path <x> <y>\r\nYour current position is (%d, %d).\r\n", ch.Room.cell.x, ch.Room.cell.y))
		return
	}

	if !ch.Room.cell.grid.isValidPosition(x, y) {
		ch.Send(fmt.Sprintf("Target (%d, %d) out of bounds.\r\n", x, y))
		return
	}

	target := ch.Room.cell.grid.grid[x][y]
	if target == nil || target.wall {
		ch.Send(fmt.Sprintf("Bad cell or obstacle at (%d, %d).\r\n", target.x, target.y))
		return
	}

	var output strings.Builder

	pathNodes := ch.Room.cell.grid.findPathAStar(ch.Room.cell, target)

	output.WriteString(fmt.Sprintf("{YPath from (%d, %d) to (%d, %d) in %d moves.{x\r\n", ch.Room.cell.x, ch.Room.cell.y, target.x, target.y, int(math.Max(0, float64(len(pathNodes)-1)))))
	for r := len(pathNodes) - 1; r >= 0; r-- {
		output.WriteString(fmt.Sprintf("{G(%d, %d){x\r\n", pathNodes[r].cell.x, pathNodes[r].cell.y))
	}

	ch.Send(output.String())
}

func (maze *MazeGrid) findPathAStar(start *MazeCell, end *MazeCell) []*MazeAStarVisit {
	visited := NewLinkedList()
	unvisited := make(map[*MazeCell]*MazeAStarVisit)

	for y := 0; y < maze.height; y++ {
		for x := 0; x < maze.width; x++ {
			unvisited[maze.grid[x][y]] = &MazeAStarVisit{gScore: 1000000, fScore: 1000000, previous: nil, cell: maze.grid[x][y]}
		}
	}

	var hScore int = maze.heuristic(start, end)
	unvisited[start] = &MazeAStarVisit{
		gScore:   0,
		fScore:   hScore,
		previous: nil,
		cell:     start,
	}

	visits := make([]*MazeAStarVisit, 0)
	if start == end {
		return visits
	}

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
				visited.Insert(unvisited[currentNode.cell])

				visits = append(visits, currentNode)
				current := currentNode.previous
				for {
					if current == nil {
						break
					}

					visits = append(visits, current)
					if current.cell == start {
						return visits
					}

					current = current.previous
				}
			} else {
				neighbours := currentNode.cell.getAdjacentCells(false, 1)
				for iter := neighbours.Head; iter != nil; iter = iter.Next {
					neighbour := iter.Value.(*MazeCell)

					if !visited.Contains(neighbour) {
						var gScore int = unvisited[currentNode.cell].gScore + 1

						_, ok := unvisited[neighbour]
						if ok {
							if gScore < unvisited[neighbour].gScore {
								unvisited[neighbour].gScore = gScore + 1 /* movement cost */
								unvisited[neighbour].fScore = gScore + maze.heuristic(currentNode.cell, end)
								unvisited[neighbour].previous = currentNode
							}
						}
					}
				}

				visited.Insert(unvisited[currentNode.cell])
				delete(unvisited, currentNode.cell)
			}
		}
	}
}
