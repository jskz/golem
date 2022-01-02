/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"time"

	"github.com/dop251/goja"
)

/*
 * Effects represent forms of enchantments that can be imbued on either the player generally
 * as a function of a spell or other game mechanic, or else an application of some attributes
 * to the player by equipment in a location.
 *
 * The Game.objectUpdate method is responsibly for expiring effects which do not have -1
 * duration.
 */
const (
	EffectTypeAffected = 0
	EffectTypeStat     = 1
	EffectTypeImmunity = 2
)

type Effect struct {
	Name       string         `json:"name"`
	EffectType int            `json:"effectType"`
	Bits       int            `json:"bits"`
	Duration   int            `json:"duration"`
	Level      int            `json:"level"`
	Location   int            `json:"location"`
	Modifier   int            `json:"modifier"`
	CreatedAt  time.Time      `json:"createdAt"`
	OnComplete *goja.Callable `json:"onComplete"`
}

/*
 * Examples of some possible effects:
 *
 * A level 50 "sanctuary" spell which was created by a cast and will last for two minutes:
 *
 * EffectType = "EffectTypeAffected",
 * Duration = 120
 * Level = 50
 * Modifier = 0
 * Location = 0
 * Bits = AFFECT_SANCTUARY
 *
 * A level 25 +2 intelligence buffing enchantment for an armor with id 50 when worn:
 *
 * EffectType = "EffectTypeStat"
 * Duration = -1
 * Level = 25
 * Modifier = 2
 * Location = WearLocationHead
 * Bits = STAT_INTELLIGENCE
 */
var AffectedFlagTable []Flag = []Flag{
	{Name: "sanctuary", Flag: AFFECT_SANCTUARY},
	{Name: "haste", Flag: AFFECT_HASTE},
	{Name: "detect_magic", Flag: AFFECT_DETECT_MAGIC},
	{Name: "fireshield", Flag: AFFECT_FIRESHIELD},
	{Name: "paralysis", Flag: AFFECT_PARALYSIS},
}

func GetAffectedFlagName(bit int) string {
	for _, flag := range AffectedFlagTable {
		if flag.Flag == bit {
			return flag.Name
		}
	}

	return "none"
}

// CreateEffect instances a new effect object; utility for scripting
func (game *Game) CreateEffect(name string, effectType int, bits int, duration int, level int, location int, modifier int, onComplete *goja.Callable) *Effect {
	return &Effect{
		Name:       name,
		EffectType: effectType,
		Bits:       bits,
		Duration:   duration,
		Level:      level,
		Location:   location,
		Modifier:   modifier,
		CreatedAt:  time.Now(),
		OnComplete: onComplete,
	}
}

func (ch *Character) AddEffect(fx *Effect) {
	switch fx.EffectType {
	case EffectTypeAffected:
		ch.Affected |= fx.Bits
	}

	ch.Effects.Insert(fx)
}

func (ch *Character) RemoveEffect(fx *Effect) {
	switch fx.EffectType {
	case EffectTypeAffected:
		ch.Affected &= ^fx.Bits
	}

	ch.Effects.Remove(fx)
}
