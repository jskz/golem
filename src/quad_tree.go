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
	Northwest *QuadTree `json:"nw"`
	Northeast *QuadTree `json:"ne"`
	Southwest *QuadTree `json:"sw"`
	Southeast *QuadTree `json:"se"`

	Boundary *Rect       `json:"boundary"`
	Nodes    *LinkedList `json:"data"`
	Capacity int         `json:"capacity"`
	Parent   *QuadTree   `json:"parent"`
}

const QuadTreeNodeMaxElements = 4

type Rect struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`

	Value interface{} `json:"value"`
}

func NewPoint(x float64, y float64, value interface{}) *Point {
	return &Point{X: x, Y: y, Value: value}
}

// Subdivide redistributes the nodes among four child trees for each subdivided rect
func (qt *QuadTree) Subdivide() bool {
	if qt.Northwest != nil {
		return false
	}

	qt.Northwest = NewQuadTree(qt.Boundary.W/2, qt.Boundary.H/2)
	qt.Northeast = NewQuadTree(qt.Boundary.W/2, qt.Boundary.H/2)
	qt.Southwest = NewQuadTree(qt.Boundary.W/2, qt.Boundary.H/2)
	qt.Southeast = NewQuadTree(qt.Boundary.W/2, qt.Boundary.H/2)

	qt.Northwest.Boundary = NewRect(qt.Boundary.X, qt.Boundary.Y, qt.Boundary.W/2, qt.Boundary.H/2)
	qt.Northeast.Boundary = NewRect(qt.Boundary.X+(qt.Boundary.W/2), qt.Boundary.Y, qt.Boundary.W/2, qt.Boundary.H/2)
	qt.Southwest.Boundary = NewRect(qt.Boundary.X, qt.Boundary.Y+(qt.Boundary.H/2), qt.Boundary.W/2, qt.Boundary.H/2)
	qt.Southeast.Boundary = NewRect(qt.Boundary.X+(qt.Boundary.W/2), qt.Boundary.Y+(qt.Boundary.H/2), qt.Boundary.W/2, qt.Boundary.H/2)

	qt.Northwest.Parent = qt
	qt.Northeast.Parent = qt
	qt.Southwest.Parent = qt
	qt.Southeast.Parent = qt

	// Repartition the nodes at this level to the appropriate child quad
	for iter := qt.Nodes.Head; iter != nil; iter = iter.Next {
		point := iter.Value.(*Point)

		qt.Nodes.Remove(point)

		if qt.Northwest.Boundary.ContainsPoint(point) {
			qt.Northwest.Nodes.Insert(point)
			break
		} else if qt.Northeast.Boundary.ContainsPoint(point) {
			qt.Northeast.Nodes.Insert(point)
			break
		} else if qt.Southwest.Boundary.ContainsPoint(point) {
			qt.Southwest.Nodes.Insert(point)
			break
		} else if qt.Southeast.Boundary.ContainsPoint(point) {
			qt.Southeast.Nodes.Insert(point)
			break
		}
	}

	return true
}

func NewRect(x float64, y float64, w float64, h float64) *Rect {
	return &Rect{X: x, Y: y, W: w, H: h}
}

func (r *Rect) Contains(x float64, y float64) bool {
	return x >= r.X && x <= r.X+r.W && y >= r.Y && y <= r.Y+r.H
}

func (r *Rect) CollidesRect(other *Rect) bool {
	minAx := r.X
	minBx := other.X
	maxAx := r.X + r.W
	maxBx := other.X + other.W

	minAy := r.Y
	minBy := other.Y
	maxAy := r.Y + r.H
	maxBy := other.Y + other.H

	return !(maxAx < minBx || minAx > maxBx || minAy > maxBy || maxAy < minBy)
}

func (r *Rect) ContainsRect(other *Rect) bool {
	return (other.X+other.W) < r.X+r.W && other.X > r.X && other.Y > r.Y && other.Y+other.H < r.Y+r.H
}

func (r *Rect) ContainsPoint(p *Point) bool {
	return r.Contains(p.X, p.Y)
}

// Insert adds a new value to the quadtree at point p
func (qt *QuadTree) Insert(p *Point) bool {
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

	if qt.Northwest.Insert(p) {
		return true
	} else if qt.Northeast.Insert(p) {
		return true
	} else if qt.Southwest.Insert(p) {
		return true
	} else if qt.Southeast.Insert(p) {
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

	// This quadtree's boundary rect does not intersect with the query rect
	if !qt.Boundary.CollidesRect(r) {
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
func NewQuadTree(width float64, height float64) *QuadTree {
	qt := &QuadTree{
		Capacity: QuadTreeNodeMaxElements,
		Nodes:    NewLinkedList(),
		Boundary: &Rect{X: 0, Y: 0, W: width, H: height},
		Parent:   nil,
	}

	return qt
}
