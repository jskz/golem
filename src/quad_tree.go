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

	Boundary *Rect       `json:"boundary"`
	Nodes    *LinkedList `json:"data"`
	Capacity int         `json:"capacity"`
}

const QuadTreeNodeMaxElements = 2

type Rect struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`

	Value interface{} `json:"value"`
}

// Subdivide redistributes the nodes among four child trees for each subdivided rect
func (qt *QuadTree) Subdivide() {
	qt.Northwest = NewQuadTree(qt.Boundary.W, qt.Boundary.H)
	qt.Northeast = NewQuadTree(qt.Boundary.W, qt.Boundary.H)
	qt.Southwest = NewQuadTree(qt.Boundary.W, qt.Boundary.H)
	qt.Southeast = NewQuadTree(qt.Boundary.W, qt.Boundary.H)
}

func (r *Rect) Contains(x int, y int) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H

}

func (r *Rect) ContainsRect(other *Rect) bool {
	return (other.X+other.W) < r.X+r.W && other.X > r.X && other.Y > r.Y && other.Y+other.H < r.Y+r.H
}

func (r *Rect) ContainsPoint(p *Point) bool {
	return r.Contains(p.X, p.Y)
}

// Insert adds a new value to the quadtree at point p
func (qt *QuadTree) Insert(p *Point, value interface{}) bool {
	if !qt.Boundary.ContainsPoint(p) {
		return false
	}

	if qt.Nodes.Count < qt.Capacity && qt.Northwest == nil {
		qt.Nodes.Insert(p)
		return true
	}

	if qt.Northwest == nil {
		qt.Subdivide()
	}

	if qt.Northwest.Insert(p, value) {
		return true
	} else if qt.Northeast.Insert(p, value) {
		return true
	} else if qt.Southwest.Insert(p, value) {
		return true
	} else if qt.Southeast.Insert(p, value) {
		return true
	}

	return false
}

// Remove removes a value from the quadtree, recursively removing nodes as necessary to "collapse" empty divisions
func (qt *QuadTree) Remove(value interface{}) {
}

// QueryRect retrieves all data within the rect defined by r
func (qt *QuadTree) QueryRect(r *Rect) []*Point {
	results := make([]*Point, 0)

	// This quadtree's boundary rect does not contain the query rect
	if !qt.Boundary.ContainsRect(r) {
		return results
	}

	for iter := qt.Nodes.Head; iter != nil; iter = iter.Next {
		p := iter.Value.(*Point)

		if r.ContainsPoint(p) {
			results = append(results, p)
		}
	}

	// This is a leaf, return results for this tree
	if qt.Northwest == nil {
		return results
	}

	// Recurse and append child tree query contents
	results = append(results, qt.Northwest.QueryRect(r)...)
	results = append(results, qt.Northeast.QueryRect(r)...)
	results = append(results, qt.Southwest.QueryRect(r)...)
	results = append(results, qt.Southeast.QueryRect(r)...)

	return results
}

// NewQuadTree creates a new quadtree instance
func NewQuadTree(width int, height int) *QuadTree {
	qt := &QuadTree{
		Capacity: QuadTreeNodeMaxElements,
		Nodes:    NewLinkedList(),
		Boundary: &Rect{X: 0, Y: 0, W: width, H: height},
	}

	return qt
}
