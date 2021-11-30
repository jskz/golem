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
	Parent   *QuadTree   `json:"parent"`
}

const QuadTreeNodeMaxElements = 2

type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type Point struct {
	X int `json:"x"`
	Y int `json:"y"`

	Value interface{} `json:"value"`
}

// Subdivide redistributes the nodes among four child trees for each subdivided rect
func (qt *QuadTree) Subdivide() {
	qt.Northwest = NewQuadTree(qt, qt.Boundary.W, qt.Boundary.H)
	qt.Northeast = NewQuadTree(qt, qt.Boundary.W, qt.Boundary.H)
	qt.Southwest = NewQuadTree(qt, qt.Boundary.W, qt.Boundary.H)
	qt.Southeast = NewQuadTree(qt, qt.Boundary.W, qt.Boundary.H)
}

func (r *Rect) Contains(x int, y int) bool {
	return float64(x) >= r.X && float64(x) <= r.X+r.W && float64(y) >= r.Y && float64(y) <= r.Y+r.H

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

// Recursively collapse quads
func (qt *QuadTree) Collapse() bool {
	// Don't further collapse the root
	if qt.Parent == nil {
		return true
	}

	// Retrieve all points within this quad
	results := qt.QueryRect(qt.Boundary)

	// If the boundary is empty, then collapse again
	if len(results) == 0 {
		return qt.Parent.Collapse()
	}

	// If there are fewer results than the capacity of a single quad, grab them and terminate here
	if len(results) < qt.Capacity {
		qt.Northwest = nil
		qt.Northeast = nil
		qt.Southwest = nil
		qt.Southeast = nil

		qt.Nodes = NewLinkedList()

		for _, p := range results {
			qt.Nodes.Insert(p)
		}

		return true
	}

	// No operation
	return true
}

// Remove removes a value from the quadtree, recursively removing nodes as necessary to "collapse" empty divisions
func (qt *QuadTree) Remove(p *Point) bool {
	// If point not in our boundary, we can't remove it
	if !qt.Boundary.ContainsPoint(p) {
		return false
	}

	// If we are in a leaf node, then remove the value
	if qt.Northwest == nil {
		qt.Nodes.Remove(p)

		// If there are other siblings in this node, no operation
		if qt.Nodes.Count > 0 {
			return true
		}

		// Recursively attempt to collapse this quad's ancestry
		return qt.Collapse()
	}

	// Try to remove from this tree's quadrants
	if qt.Northwest.Remove(p) {
		return true
	} else if qt.Northeast.Remove(p) {
		return true
	} else if qt.Southwest.Remove(p) {
		return true
	} else if qt.Southeast.Remove(p) {
		return true
	}

	return false
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
func NewQuadTree(parent *QuadTree, width float64, height float64) *QuadTree {
	qt := &QuadTree{
		Capacity: QuadTreeNodeMaxElements,
		Nodes:    NewLinkedList(),
		Boundary: &Rect{X: 0, Y: 0, W: width, H: height},
		Parent:   parent,
	}

	return qt
}
