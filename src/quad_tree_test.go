/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"math/rand"
	"testing"
	"time"
)

type ExampleStructure struct {
	Id int
}

// Generate a set of n points that fall within dimensions w and h
func CreateTestPoints(n int, w int, h int) []*Point {
	points := []*Point{}

	r := rand.New(rand.NewSource(time.Now().Unix()))

	for i := 0; i < n; i++ {
		p := &Point{X: float64(r.Intn(w)), Y: float64(r.Intn(h)), Value: &ExampleStructure{Id: i}}
		points = append(points, p)
	}

	return points
}

func TestQuadTreeInsert(t *testing.T) {
	qt := NewQuadTree(128, 128)
	examplePoints := CreateTestPoints(32, 128, 128)

	for _, p := range examplePoints {
		qt.Insert(p)
	}

	results := qt.QueryRect(NewRect(0, 0, 128, 128))
	if len(results) != len(examplePoints) {
		t.Errorf("Root quadtree boundary contained %d points, expected %d.", len(results), len(examplePoints))
	}

	matches := make([]interface{}, 0)
	for _, p := range examplePoints {
		for _, match := range results {
			if match == p {
				matches = append(matches, match)
			}
		}
	}

	if len(matches) != len(examplePoints) {
		t.Errorf("Found %d of the example data, expected %d.", len(matches), len(examplePoints))
	}
}

func TestQuadTreeRemove(t *testing.T) {
	qt := NewQuadTree(128, 128)
	examplePoints := CreateTestPoints(32, 128, 128)

	for _, p := range examplePoints {
		qt.Insert(p)
	}

	// Remove a single point
	var head *Point

	head, examplePoints = examplePoints[0], examplePoints[1:]
	qt.Remove(head)

	results := qt.QueryRect(NewRect(0, 0, 128, 128))
	if len(results) != 31 {
		t.Errorf("Root quadtree boundary contained %d points, expected 31.", len(results))
	}

	// Remove the rest of the points
	for _, removal := range examplePoints {
		qt.Remove(removal)
	}

	results = qt.QueryRect(NewRect(0, 0, 128, 128))
	if len(results) > 0 {
		t.Errorf("Root quadtree boundary contained %d points, expected 0.", len(results))
	}

	// Is the tree now a leaf?
	if qt.Northwest != nil {
		t.Error("Tree still had child quadrants after removing all nodes")
	}

	// Is the leaf empty?
	if qt.Nodes.Count != 0 {
		t.Errorf("Root quadtree expected empty, had %d nodes.\r\n", qt.Nodes.Count)
	}
}
