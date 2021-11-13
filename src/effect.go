/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "time"

/*
 * Effects represent forms of enchantments that can be imbued on either the player generally
 * as a function of a spell or other game mechanic, or else an application of some attributes
 * to the player by equipment in a location.
 *
 * The Game.objectUpdate method is responsibly for expiring effects which do not have -1
 * duration.
 */
const (
	EffectTypeAffected  = 0
	EffectTypeEquipment = 1
	EffectTypeImmunity  = 2
)

type Effect struct {
	EffectType int       `json:"effectType"`
	Bits       int       `json:"bits"`
	Duration   int       `json:"duration"`
	Level      int       `json:"level"`
	Location   int       `json:"location"`
	Modifier   int       `json:"modifier"`
	CreatedAt  time.Time `json:"createdAt"`
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
 * A level 25 +2 intelligence buffing enchantment for a helmet when worn:
 *
 * EffectType = "EffectTypeEquipment"
 * Duration = -1
 * Level = 25
 * Modifier = 2
 * Location = WearLocationHead
 * Bits = MOD_INTELLIGENCE
 */

// Instance a new effect object, handy for scripting
func (game *Game) CreateEffect(effectType int, bits int, duration int, level int, location int, modifier int) *Effect {
	return &Effect{
		EffectType: effectType,
		Bits:       bits,
		Duration:   duration,
		Level:      level,
		Location:   location,
		Modifier:   modifier,
		CreatedAt:  time.Now(),
	}
}
