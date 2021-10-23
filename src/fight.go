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

func (game *Game) createBlood(intensity int) *ObjectInstance {
	obj := &ObjectInstance{game: game}

	obj.parentId = 1
	obj.description = fmt.Sprintf("{rA puddle of blood has spilled here.{x")
	obj.shortDescription = fmt.Sprintf("a puddle of blood")
	obj.longDescription = fmt.Sprintf("{rThere is a puddle of blood here.")
	obj.name = "blood puddle"
	obj.itemType = ItemTypeNone
	obj.createdAt = time.Now()
	obj.flags = ITEM_DECAYS | ITEM_DECAY_SILENTLY
	obj.ttl = 5

	return obj
}

func (game *Game) createCorpse(ch *Character) *ObjectInstance {
	obj := &ObjectInstance{game: game}

	obj.parentId = 1
	obj.description = fmt.Sprintf("The slain corpse of %s.", ch.GetShortDescription(ch))
	obj.shortDescription = fmt.Sprintf("the corpse of %s", ch.GetShortDescription(ch))
	obj.longDescription = fmt.Sprintf("The corpse of %s is lying here.", ch.GetShortDescription(ch))
	obj.name = fmt.Sprintf("corpse %s", ch.name)
	obj.itemType = "container"
	obj.createdAt = time.Now()
	obj.flags = ITEM_DECAYS
	obj.ttl = 20

	if ch.flags&CHAR_IS_PLAYER == 0 {
		obj.contents = NewLinkedList()
		obj.contents.Tail = ch.inventory.Tail
		obj.contents.Head = ch.inventory.Head
		obj.contents.Count = ch.inventory.Count

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
			for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
				character := iter.Value.(*Character)
				if character != ch && character != target {
					character.Send(fmt.Sprintf("{G%s{G hits %s{G for %d damage.{x\r\n",
						ch.GetShortDescriptionUpper(character),
						target.GetShortDescription(character),
						amount))
				}
			}
		}

		ch.Send(fmt.Sprintf("{GYou hit %s{G for %d damage.{x\r\n", target.GetShortDescription(ch), amount))
		target.Send(fmt.Sprintf("{Y%s{Y hits you for %d damage.{x\r\n", ch.GetShortDescriptionUpper(target), amount))
	}

	target.health -= amount

	if target.health > target.maxHealth {
		target.health = target.maxHealth
	}

	if target.health <= 0 {
		if target.Room != nil {
			room := target.Room

			corpse := game.createCorpse(target)
			room.removeCharacter(target)
			room.addObject(corpse)

			blood := game.createBlood(1)
			room.addObject(blood)

			target.Fighting = nil
			target.Combat = nil

			for iter := room.Characters.Head; iter != nil; iter = iter.Next {
				character := iter.Value.(*Character)
				character.Send(fmt.Sprintf("{R%s{R has been slain!{x\r\n", target.GetShortDescriptionUpper(character)))

				if character.Fighting.IsEqual(target) {
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
				exp := int(target.experience)
				if ch != nil {
					if ch.Group != nil {
						groupExperience := exp / ch.Group.Count

						for iter := ch.Group.Head; iter != nil; iter = iter.Next {
							gch := iter.Value.(*Character)

							gch.gainExperience(groupExperience)
						}
					} else {
						ch.gainExperience(int(exp))
					}
				}

				game.Characters.Remove(target)
				target = nil
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

	var exits []*Exit = make([]*Exit, 0)

	for _, exit := range ch.Room.Exit {
		if exit.To != nil && exit.Flags&EXIT_CLOSED == 0 {
			exits = append(exits, exit)
		}
	}

	if rand.Intn(10) < 7 {
		ch.Send("{RYou panic and attempt to flee, but can't get away!{x\r\n")

		/* Announce player's failed flee attempt to others in the room */
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch != ch {
				output := fmt.Sprintf("\r\n{R%s{R panics and attempts to flee, but can't get away!{x\r\n", ch.GetShortDescriptionUpper(rch))
				rch.Send(output)
			}
		}

		return
	}

	var choice int = rand.Intn(len(exits))
	var chosenEscape *Exit = exits[choice]

	ch.Send(fmt.Sprintf("{RYou panic and flee %s!{x\r\n", ExitName[chosenEscape.Direction]))

	/* Announce player's departure to all other players in the current room */
	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			/* If they were fighting this player, then this is a good time to stop their participation in the fight and let it dispose */
			if rch.Fighting == ch {
				rch.Fighting = nil
				rch.Combat = nil
			}

			output := fmt.Sprintf("\r\n{R%s{R has fled %s!{x\r\n", ch.GetShortDescriptionUpper(rch), ExitName[chosenEscape.Direction])
			rch.Send(output)
		}
	}

	ch.Fighting = nil
	ch.Combat = nil

	ch.Room.Characters.Remove(ch)
	ch.Room = chosenEscape.To
	chosenEscape.To.Characters.Insert(ch)

	/* Announce player's arrival to all other players in the new room */
	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			output := fmt.Sprintf("\r\n{W%s{W arrives from %s.{x\r\n", ch.GetShortDescriptionUpper(rch), ExitName[ReverseDirection[chosenEscape.Direction]])
			rch.Send(output)
		}
	}

	do_look(ch, "")
}

func do_kill(ch *Character, arguments string) {
	if ch.Room == nil {
		return
	}

	if ch.Room.flags&ROOM_SAFE != 0 {
		ch.Send("{WYou cannot do that here.{x\r\n")
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

	var target *Character = ch.FindCharacterInRoom(arguments)

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
		target.Combat = combat
	}

	ch.Send(fmt.Sprintf("\r\n{GYou begin attacking %s{G!{x\r\n", target.GetShortDescription(ch)))
	target.Send(fmt.Sprintf("\r\n{G%s{G begins attacking you!{x\r\n", ch.GetShortDescriptionUpper(target)))

	if ch.Room != nil && target.Room != nil && target.Room == ch.Room {
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			character := iter.Value.(*Character)
			if character != ch && character != target {
				character.Send(fmt.Sprintf("{G%s{G begins attacking %s{G!{x\r\n",
					ch.GetShortDescriptionUpper(character),
					target.GetShortDescription(character)))
			}
		}
	}
}
