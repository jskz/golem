/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

/*
 * Quadtree will give us a primitive for spatially indexing pointers of game objects; our primary use case for
 * this data structure will be in the world map where we will lean on it for lookups instead of using the usual
 * room character/object/entity lists for a given room.
 */
type QuadTree struct {
	Northwest *QuadTree
	Northeast *QuadTree
	Southwest *QuadTree
	Southeast *QuadTree
}

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Insert adds a new value to the quadtree at point p
func (qt *QuadTree) Insert(p Point, value interface{}) {
}

// Remove removes a value from the quadtree, recursively removing nodes as necessary to "collapse" empty divisions
func (qt *QuadTree) Remove(value interface{}) {
}

// QueryRect retrieves all data within the rect defined by r
func (qt *QuadTree) QueryRect(r Rect) []interface{} {
	return nil
}
