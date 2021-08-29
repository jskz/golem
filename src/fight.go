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
	"strings"
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

	if display {
		ch.Send(fmt.Sprintf("You hit %s for %d damage.\r\n", target.getShortDescription(), amount))
		target.Send(fmt.Sprintf("%s hit you for %d damage.\r\n", ch.getShortDescription(), amount))
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
		}
	}
}

func do_kill(ch *Character, arguments string) {
	if ch.room == nil {
		return
	}

	if ch.fighting != nil {
		ch.Send("You are already fighting somebody else!\r\n")
		return
	}

	if len(arguments) < 1 {
		ch.Send("Attack who?\r\n")
		return
	}

	var target *Character = nil

	for rch := range ch.room.characters {
		if strings.Contains(rch.name, arguments) {
			target = rch
		}
	}

	if target == ch || target == nil {
		ch.Send("No such target.  Attack who?\r\n")
		return
	}

	combat := &Combat{}
	combat.startedAt = time.Now()
	combat.participants = []*Character{ch, target}
	ch.client.game.fights.Insert(combat)

	ch.fighting = target
	ch.Send(fmt.Sprintf("\r\n{RYou begin attacking %s{R!{x\r\n", target.getShortDescription()))
}
