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
	"math/rand"
)

/*
 * Perlin noise generator; we'll expose this to the scripting API as a utility method for helping to generate
 * "islands" or other gradient-based planes.
 *
 * https://rtouti.github.io/graphics/perlin-noise-algorithm provided the helpful reference implementation and
 * explanatory article.
 *
 * We may generalize this yet ;)
 */
type PerlinVector2D struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

func (v PerlinVector2D) DotProduct(other PerlinVector2D) float64 {
	return v.X*other.X + v.Y*other.Y
}

func GeneratePermutation() []int {
	P := make([]int, 256)

	var x int

	for x = 0; x < 256; x++ {
		P[x] = x
	}

	/* shuffle bytes in-place */
	for i := range P {
		j := rand.Intn(i + 1)

		P[i], P[j] = P[j], P[i]
	}

	for x = 0; x < 256; x++ {
		P = append(P, P[x])
	}

	return P
}

func GetConstantVector(value byte) PerlinVector2D {
	b := value & 3

	if b == 0 {
		return PerlinVector2D{1.0, 1.0}
	} else if b == 1 {
		return PerlinVector2D{-1.0, 1.0}
	} else if b == 2 {
		return PerlinVector2D{-1.0, -1.0}
	}

	return PerlinVector2D{1.0, -1.0}
}

func Noise2D(x float64, y float64, P []byte) float64 {
	X := byte(x) & 255
	Y := byte(y) & 255

	xf := x - math.Floor(x)
	yf := y - math.Floor(y)

	c10 := PerlinVector2D{X: xf - 1.0, Y: yf - 1.0}
	c00 := PerlinVector2D{X: xf, Y: yf - 1.0}
	c11 := PerlinVector2D{X: xf - 1.0, Y: yf}
	c01 := PerlinVector2D{X: xf, Y: yf}

	vc10 := P[P[X+1]+Y+1]
	vc00 := P[P[X]+Y+1]
	vc11 := P[P[X+1]+Y]
	vc01 := P[P[X]+Y]

	dp10 := c10.DotProduct(GetConstantVector(vc10))
	dp00 := c00.DotProduct(GetConstantVector(vc00))
	dp11 := c11.DotProduct(GetConstantVector(vc11))
	dp01 := c01.DotProduct(GetConstantVector(vc01))

	u := Fade(xf)
	v := Fade(yf)

	return Lerp2D(u, Lerp2D(v, dp01, dp00), Lerp2D(v, dp11, dp10))
}

func GeneratePerlinNoise(w int, h int)
