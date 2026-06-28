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
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
	"unicode"
)

const simpleHTTPTimeout = 5 * time.Second

var simpleHTTPClient = &http.Client{
	Timeout: simpleHTTPTimeout,
}

type Flag struct {
	Name string `json:"name"`
	Flag int    `json:"flag"`
}

type CharacterStatValue struct {
	Modified int `json:"modified"`
	Base     int `json:"base"`
}

func FlagNames(flags int, table []Flag) string {
	var output strings.Builder
	remaining := flags

	if flags == 0 {
		return "none"
	}

	for _, flag := range table {
		if flags&flag.Flag == 0 {
			continue
		}

		output.WriteString(fmt.Sprintf("%s ", flag.Name))
		remaining &= ^flag.Flag
	}

	if remaining != 0 {
		output.WriteString(fmt.Sprintf("unknown(%d) ", remaining))
	}

	if output.Len() == 0 {
		return "none"
	}

	return strings.TrimRight(output.String(), " ")
}

func CharacterFlagNames(flags int) string {
	return FlagNames(flags, CharacterFlagTable)
}

func AffectedFlagNames(flags int) string {
	return FlagNames(flags, AffectedFlagTable)
}

func PositionName(position int) string {
	switch position {
	case PositionDead:
		return "dead"
	case PositionStunned:
		return "stunned"
	case PositionSleeping:
		return "sleeping"
	case PositionResting:
		return "resting"
	case PositionSitting:
		return "sitting"
	case PositionFighting:
		return "fighting"
	case PositionStanding:
		return "standing"
	default:
		return fmt.Sprintf("unknown(%d)", position)
	}
}

func StatName(stat int) string {
	name, ok := StatNameTable[stat]
	if !ok {
		return "unknown"
	}

	return name
}

func CharacterStat(ch *Character, stat int) CharacterStatValue {
	if ch == nil {
		return CharacterStatValue{}
	}

	if stat < 0 || stat >= len(ch.Stats) {
		return CharacterStatValue{}
	}

	modified, base := ch.GetStat(stat)
	return CharacterStatValue{Modified: modified, Base: base}
}

func CharacterName(ch *Character) string {
	if ch == nil {
		return "none"
	}

	if ch.Name != "" {
		return ch.Name
	}

	if ch.ShortDescription != "" {
		return ch.ShortDescription
	}

	return "unnamed"
}

func CharacterTypeName(ch *Character) string {
	if ch != nil && ch.Flags&CHAR_IS_PLAYER != 0 {
		return "PC"
	}

	return "NPC"
}

func CharacterLocationName(ch *Character) string {
	if ch == nil || ch.Room == nil {
		return "none"
	}

	room := ch.Room
	roomName := room.Name
	if roomName == "" {
		roomName = "unnamed room"
	}

	if room.Flags&ROOM_PLANAR != 0 && room.Plane != nil {
		return fmt.Sprintf("%s (#%d, plane #%d @ %d,%d,%d)", roomName, room.Id, room.Plane.Id, room.X, room.Y, room.Z)
	}

	return fmt.Sprintf("%s (#%d)", roomName, room.Id)
}

func EffectDurationDescription(fx *Effect) string {
	if fx == nil {
		return "none"
	}

	if fx.Duration == EffectDurationPermanent {
		return "permanent"
	}

	remaining := int(math.Max(0, time.Until(fx.CreatedAt.Add(time.Duration(fx.Duration)*time.Second)).Seconds()))
	return fmt.Sprintf("%d seconds", remaining)
}

func EffectDescription(fx *Effect) string {
	if fx == nil {
		return "none"
	}

	switch fx.EffectType {
	case EffectTypeStat:
		return fmt.Sprintf("%s level %d modifies %s by %d for %s",
			fx.Name,
			fx.Level,
			StatName(fx.Location),
			fx.Modifier,
			EffectDurationDescription(fx))
	case EffectTypeAffected:
		return fmt.Sprintf("%s level %d applies %s for %s",
			fx.Name,
			fx.Level,
			AffectedFlagNames(fx.Bits),
			EffectDurationDescription(fx))
	case EffectTypeImmunity:
		return fmt.Sprintf("%s level %d grants immunity bits %d for %s",
			fx.Name,
			fx.Level,
			fx.Bits,
			EffectDurationDescription(fx))
	default:
		return fmt.Sprintf("%s level %d type %d for %s",
			fx.Name,
			fx.Level,
			fx.EffectType,
			EffectDurationDescription(fx))
	}
}

func SimpleGET(url string) (string, error) {
	resp, err := simpleHTTPClient.Get(url)
	if err != nil {
		return "", err
	}

	return readSimpleHTTPResponse(resp)
}

func SimplePOST(url string, data string) (string, error) {
	resp, err := simpleHTTPClient.Post(url,
		"application/json",
		bytes.NewBuffer([]byte(data)))
	if err != nil {
		return "", err
	}

	return readSimpleHTTPResponse(resp)
}

func readSimpleHTTPResponse(resp *http.Response) (string, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http request failed with status %s", resp.Status)
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

func resourcePercentage(current int, maximum int) int {
	if maximum <= 0 || current <= 0 {
		return 0
	}

	if current >= maximum {
		return 100
	}

	return int(float64(current) * 100 / float64(maximum))
}
