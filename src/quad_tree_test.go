package main

import (
	"math/rand"
	"testing"
)

type ExampleStructure struct {
	Id int
}

// Generate a set of n points that fall within dimensions w and h
func CreateTestPoints(n int, w int, h int) []*Point {
	points := []*Point{}

	for i := 0; i < n; i++ {
		p := &Point{X: float64(rand.Intn(w)), Y: float64(rand.Intn(h)), Value: &ExampleStructure{Id: i}}
		points = append(points, p)
	}

	return points
}

func TestQuadTreeInsert(t *testing.T) {
	examplePoints := CreateTestPoints(32, 128, 128)

	qt := NewQuadTree(128, 128)
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
}
