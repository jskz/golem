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
	"math/rand"
	"time"
)

type Combat struct {
	StartedAt    time.Time    `json:"startedAt"`
	Room         *Room        `json:"room"`
	Participants []*Character `json:"participants"`
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
	obj.shortDescription = fmt.Sprintf("the corpse of %s", ch.getShortDescription(ch))
	obj.longDescription = fmt.Sprintf("The corpse of %s is lying here.", ch.getShortDescription(ch))
	obj.name = fmt.Sprintf("corpse %s", ch.name)
	obj.itemType = "container"

	if ch.flags&CHAR_IS_PLAYER == 0 {
		obj.contents = ch.inventory

		ch.inventory = NewLinkedList()
	} else {
		obj.contents = NewLinkedList()
	}

	return obj
}

func (game *Game) Damage(ch *Character, target *Character, display bool, amount int, damageType int) bool {
	if target == nil {
		return false
	}

	if display && ch != nil {
		if ch.Room != nil && target.Room != nil && target.Room == ch.Room {
			for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
				character := iter.Value.(*Character)
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
	if target.health <= 0 {
		if target.Room != nil {
			room := target.Room

			corpse := game.createCorpse(target)
			room.removeCharacter(target)
			room.addObject(corpse)

			target.Fighting = nil
			target.Combat = nil

			for iter := room.characters.Head; iter != nil; iter = iter.Next {
				character := iter.Value.(*Character)
				character.Send(fmt.Sprintf("{R%s{R has been slain!{x\r\n", target.getShortDescriptionUpper(character)))

				if character.Fighting == target {
					character.Fighting = nil
				}
			}

			if target.flags&CHAR_IS_PLAYER != 0 {
				target.Send("{RYou have been slain!{D\r\n")
				target.Send(string(Config.death))
				target.Send("{x\r\n")

				limbo, err := game.LoadRoomIndex(RoomLimbo)
				if err != nil {
					return true
				}

				limbo.addCharacter(target)
				target.health = target.maxHealth / 8
				do_look(target, "")
			} else {
				exp := target.experience
				if ch != nil {
					ch.gainExperience(int(exp))
				}

				game.characters.Remove(target)
			}
		}
	}

	return true
}

func (game *Game) combatUpdate() {
	game.InvokeNamedEventHandlersWithContextAndArguments("combatUpdate", game.vm.ToValue(game))
}

func (game *Game) DisposeCombat(combat *Combat) {
	for _, vch := range combat.Participants {
		vch.Combat = nil
		vch.Fighting = nil
	}

	game.Fights.Remove(combat)
}

func do_flee(ch *Character, arguments string) {
	if ch.Room == nil {
		return
	}

	if ch.Fighting == nil {
		ch.Send("{RYou can't flee while not fighting.{x\r\n")
		return
	}

	if ch.casting != nil {
		ch.Send("{RYou are too concentrated on casting a magical spell to flee from combat.{x\r\n")
		return
	}

	/* TODO: other logic/affects preventing a player from fleeing */
	var exits []*Exit = make([]*Exit, 0)

	for _, exit := range ch.Room.exit {
		if exit.to != nil {
			exits = append(exits, exit)
		}
	}

	if rand.Intn(10) < 7 {
		ch.Send("{RYou panic and attempt to flee, but can't get away!{x\r\n")

		/* Announce player's failed flee attempt to others in the room */
		for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch != ch {
				output := fmt.Sprintf("\r\n{R%s{R panics and attempts to flee, but can't get away!{x\r\n", ch.getShortDescriptionUpper(rch))
				rch.Send(output)
			}
		}

		return
	}

	var choice int = rand.Intn(len(exits))
	var chosenEscape *Exit = exits[choice]

	ch.Send(fmt.Sprintf("{RYou panic and flee %s!{x\r\n", ExitName[chosenEscape.direction]))

	/* Announce player's departure to all other players in the current room */
	for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			output := fmt.Sprintf("\r\n{R%s{R has fled %s!{x\r\n", ch.getShortDescriptionUpper(rch), ExitName[chosenEscape.direction])
			rch.Send(output)
		}
	}

	ch.Fighting = nil
	ch.Combat = nil

	ch.Room.characters.Remove(ch)
	ch.Room = chosenEscape.to
	chosenEscape.to.characters.Insert(ch)

	/* Announce player's arrival to all other players in the new room */
	for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			output := fmt.Sprintf("\r\n{W%s{W arrives from %s.{x\r\n", ch.getShortDescriptionUpper(rch), ExitName[ReverseDirection[chosenEscape.direction]])
			rch.Send(output)
		}
	}

	do_look(ch, "")
}

func do_kill(ch *Character, arguments string) {
	if ch.Room == nil {
		return
	}

	if ch.Fighting != nil {
		ch.Send("You are already fighting somebody else!\r\n")
		return
	}

	if len(arguments) < 1 {
		ch.Send("Attack who?\r\n")
		return
	}

	var target *Character = ch.findCharacterInRoom(arguments)

	if target == ch || target == nil {
		ch.Send("No such target.  Attack who?\r\n")
		return
	}

	combat := &Combat{}
	combat.StartedAt = time.Now()
	combat.Room = ch.Room
	combat.Participants = []*Character{ch, target}
	ch.game.Fights.Insert(combat)

	ch.Fighting = target

	if target.Fighting == nil {
		target.Fighting = ch
	}

	if target.Combat == nil {
		target.Combat = combat
	}

	ch.Send(fmt.Sprintf("\r\n{GYou begin attacking %s{G!{x\r\n", target.getShortDescription(ch)))
	target.Send(fmt.Sprintf("\r\n{G%s{G begins attacking you!{x\r\n", ch.getShortDescriptionUpper(target)))

	if ch.Room != nil && target.Room != nil && target.Room == ch.Room {
		for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
			character := iter.Value.(*Character)
			if character != ch && character != target {
				character.Send(fmt.Sprintf("{G%s{G begins attacking %s{G!{x\r\n",
					ch.getShortDescriptionUpper(character),
					target.getShortDescription(character)))
			}
		}
	}
}
