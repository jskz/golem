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
	"math/rand"
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

func (game *Game) createCorpse(ch *Character) *ObjectInstance {
	obj := &ObjectInstance{}

	obj.id = 1
	obj.description = fmt.Sprintf("The slain corpse of %s.", ch.getShortDescription(ch))
	obj.longDescription = fmt.Sprintf("The corpse of %s is lying here.", ch.getShortDescription(ch))
	obj.name = fmt.Sprintf("corpse %s", ch.name)

	return obj
}

func (game *Game) damage(ch *Character, target *Character, display bool, amount int, damageType int) bool {
	if ch == nil || target == nil {
		return false
	}

	if display {
		if ch.room != nil && target.room != nil && target.room == ch.room {
			for character := range ch.room.characters {
				if character != ch && character != target {
					character.Send(fmt.Sprintf("{G%s{G hits %s{G for %d damage.{x\r\n",
						ch.getShortDescriptionUpper(character),
						target.getShortDescription(character),
						amount))
				}
			}
		}

		ch.Send(fmt.Sprintf("{GYou hit %s{G for %d damage.{x\r\n", target.getShortDescription(ch), amount))
		target.Send(fmt.Sprintf("{Y%s{Y hits you for %d damage.{x\r\n", ch.getShortDescriptionUpper(target), amount))
	}

	target.health -= amount
	if target.health < 0 {
		if target.room != nil {
			room := target.room

			corpse := game.createCorpse(target)
			room.removeCharacter(target)
			room.addObject(corpse)

			for character := range room.characters {
				character.Send(fmt.Sprintf("{R%s{R has been slain!{x\r\n", target.getShortDescriptionUpper(character)))
			}

			if target.flags&CHAR_IS_PLAYER != 0 {
				target.Send("{RYou have been slain!{x\r\n")

				limbo, err := game.LoadRoomIndex(RoomLimbo)
				if err != nil {
					return true
				}

				limbo.addCharacter(target)
				target.health = target.maxHealth / 2
			} else {
				exp := target.experience
				ch.gainExperience(int(exp))
			}
		}
	}

	return true
}

func (game *Game) combatUpdate() {
	for iter := game.fights.head; iter != nil; iter = iter.next {
		combat := iter.value.(*Combat)

		var found bool = false
		for _, vch := range combat.participants {
			if vch.fighting == nil {
				log.Printf("Participant target not currently fighting.\r\n")
				continue
			}

			if vch.room == nil || vch.fighting.room == nil || vch.room != vch.fighting.room {
				log.Printf("Some participants are in nil or mismatched rooms and not considered this round.\r\n")
				continue
			}

			found = true

			/* Determine number of attacks - agility, haste/slow spells, etc */
			var attackerRounds int = 1
			var dexterityBonusRounds int = int(float64((vch.dexterity - 10) / 4))

			attackerRounds += dexterityBonusRounds

			for round := 0; round < attackerRounds; round++ {
				damage := 0

				/* No weapon equipped, calculate unarmed damage with strength and skill */
				damage = rand.Intn(2)

				/* Modify with attributes */
				damage += rand.Intn(vch.strength / 3)

				/* Evasion check */
				game.damage(vch, vch.fighting, true, damage, DamageTypeBash)
			}
		}

		if !found {
			log.Printf("Discarding fight without active participants.\r\n")

			game.disposeCombat(combat)
			break
		}
	}
}

func (game *Game) disposeCombat(combat *Combat) {
	for _, vch := range combat.participants {
		vch.combat = nil
		vch.fighting = nil
	}

	game.fights.Remove(combat)
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

	if target.fighting == nil {
		target.fighting = ch
	}

	if target.combat == nil {
		target.combat = combat
	}

	ch.Send(fmt.Sprintf("\r\n{GYou begin attacking %s{G!{x\r\n", target.getShortDescription(ch)))
	target.Send(fmt.Sprintf("\r\n{G%s{G begins attacking you!{x\r\n", ch.getShortDescriptionUpper(target)))

	if ch.room != nil && target.room != nil && target.room == ch.room {
		for character := range ch.room.characters {
			if character != ch && character != target {
				character.Send(fmt.Sprintf("{G%s{G begins attacking %s{G!{x\r\n",
					ch.getShortDescriptionUpper(character),
					target.getShortDescription(character)))
			}
		}
	}
}
