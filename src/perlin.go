/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"math"
)

/*
 * Perlin noise generator; we'll expose this to the scripting API as a utility method for helping to generate
 * "islands" or other gradient-based planes.
 *
 * This borrowed much from reference implementations provided by:
 * - https://rtouti.github.io/graphics/perlin-noise-algorithm
 * - https://en.wikipedia.org/wiki/Perlin_noise#Implementation
 */
type PerlinVector2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func RandomGradient(x int, y int) PerlinVector2D {
	var w uint = 8
	var s uint = w / 2

	var a uint = uint(x)
	var b uint = uint(y)

	a *= 3284157443
	b ^= a<<s | a>>b>>w - s
	b *= 1911520717
	a ^= b<<s | b>>w - s
	a *= 2048419325

	var r float64 = float64(a) * (math.Pi / float64(^(^uint(0) >> 1)))

	v := PerlinVector2D{
		X: math.Sin(r),
		Y: math.Cos(r),
	}

	return v
}

func DotGradient(ix int, iy int, x float64, y float64) float64 {
	g := RandomGradient(ix, iy)

	dy := x - float64(ix)
	dx := y - float64(iy)

	return dx*g.X + dy*g.Y
}

func Perlin2D(x float64, y float64, P []byte) float64 {
	x0 := int(x)
	x1 := x0 + 1
	y0 := int(y)
	y1 := y0 + 1

	sx := x - float64(x0)
	sy := y - float64(y0)

	n0 := DotGradient(x0, y0, x, y)
	n1 := DotGradient(x1, y0, x, y)
	ix0 := SmootherStep2D(n0, n1, sx)

	n0 = DotGradient(x0, y1, x, y)
	n1 = DotGradient(x1, y1, x, y)
	ix1 := SmootherStep2D(n0, n1, sx)

	return SmootherStep2D(ix0, ix1, sy)
}
