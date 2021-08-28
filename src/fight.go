/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"fmt"
	"log"
	"time"
)

type Combat struct {
	startedAt    time.Time
	participants []*Character
}

const (
	DamageTypeBash   = 0
	DamageTypeSlash  = 1
	DamageTypeStab   = 2
	DamageTypeExotic = 3
)

func (game *Game) damage(ch *Character, target *Character, display bool, amount int, damageType int) bool {
	if ch == nil || target == nil {
		return false
	}

	return true
}

func (game *Game) combatUpdate() {
	for iter := game.fights.head; iter != nil; iter = iter.next {
		combat := iter.value.(*Combat)

		td := time.Since(combat.startedAt)
		log.Printf("Combat for %d seconds, calculating damage for round.\r\n", int(td.Seconds()))

		for _, vch := range combat.participants {
			if vch.fighting == nil {
				log.Printf("Participant target not currently fighting.\r\n")
				continue
			}

			damage := 0

			game.damage(vch, vch.fighting, true, damage, DamageTypeBash)
			vch.Send(fmt.Sprintf("You did %d damage in a combat round.\r\n", damage))
		}
	}
}
