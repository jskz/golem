/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"bytes"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"unicode"
)

type Flag struct {
	Name string `json:"name"`
	Flag int    `json:"flag"`
}

func SimpleGET(url string, data string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func SimplePOST(url string, data string) (string, error) {
	resp, err := http.Post(url,
		"application/json",
		bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func OneArgument(input string) (string, string) {
	var args string = strings.TrimSpace(input)
	var buf strings.Builder
	var quoted bool = false
	var end int = len(args)

	for index, r := range args {
		if r == '\'' || r == '"' {
			if quoted {
				end = index
				break
			}

			quoted = true
		} else {
			if r != ' ' || quoted {
				buf.WriteRune(unicode.ToLower(r))
			} else if r == ' ' && !quoted {
				end = index
				break
			}
		}
	}

	if quoted && end < len(args) {
		end++
	}

	return buf.String(), strings.TrimLeft(args[end:], " ")
}

func Fade(t float64) float64 {
	return ((6*t-15)*t + 10) * t * t * t
}

func Lerp2D(s float64, e float64, t float64) float64 {
	return s + (e-s)*t
}

func SmootherStep2D(a0 float64, a1 float64, w float64) float64 {
	return (a1-a0)*((w*(w*6.0-15.0)+10.0)*w*w*w) + a0
}

func Distance2D(x float64, y float64, x2 float64, y2 float64, a float64, b float64) int {
	return int(math.Sqrt(((((x2 - x) * (x2 - x)) / (a * a)) + ((y2-y)*(y2-y))/(b*b))))
}

func Angle2D(x float64, y float64, x2 float64, y2 float64) int {
	dy := y2 - y
	dx := x2 - x

	radians := math.Atan2(-dx, dy)
	degrees := radians * (180 / math.Pi)

	if degrees <= 0 {
		degrees += 360
	}

	return int(degrees)
}

func AngleToDirection(angle int) int {
	if angle < 45 {
		return DirectionNorth
	} else if angle < 90 {
		return DirectionNortheast
	} else if angle < 135 {
		return DirectionEast
	} else if angle < 180 {
		return DirectionSoutheast
	} else if angle < 225 {
		return DirectionSouth
	} else if angle < 270 {
		return DirectionSouthwest
	} else if angle < 315 {
		return DirectionWest
	} else if angle < 360 {
		return DirectionNorthwest
	}

	return DirectionNorth
}
