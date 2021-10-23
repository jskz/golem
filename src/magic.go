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
	"strings"
	"time"

	"github.com/dop251/goja"
)

type CastingContext struct {
	casting    *Skill
	arguments  string
	startedAt  time.Time
	complexity int
}

func (ch *Character) onCastingUpdate() {
	finishedCastingAt := ch.casting.startedAt.Add(time.Duration(ch.casting.complexity) * time.Second)
	sinceCastingFinished := time.Since(finishedCastingAt)

	if sinceCastingFinished.Seconds() >= 0 {
		ch.Send("\r\n{WYou finish casting the magic spell.{x\r\n")

		if ch.Room != nil {
			for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
				rch := iter.Value.(*Character)

				if !rch.IsEqual(ch) {
					rch.Send(fmt.Sprintf("\r\n{W%s{W finishes casting their magic spell.{W{x.\r\n", ch.GetShortDescriptionUpper(rch)))
				}
			}
		}

		if ch.casting.casting.handler != nil {
			fn := *ch.casting.casting.handler

			fn(ch.game.vm.ToValue(ch.casting), ch.game.vm.ToValue(ch), ch.game.vm.ToValue(ch.casting.arguments))
		}

		ch.casting = nil
	}
}

func (game *Game) RegisterSpellHandler(name string, fn goja.Callable) goja.Value {
	spell := game.FindSkillByName(name)
	if spell == nil || spell.skillType != SkillTypeSpell {
		return game.vm.ToValue(nil)
	}

	spell.handler = &fn
	return game.vm.ToValue(spell)
}

func do_cast(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Cast what?\r\n")
		return
	}

	arg, arguments := oneArgument(arguments)

	if ch.casting != nil {
		ch.Send("You are already in the middle of casting another spell!\r\n")
		return
	}

	if ch.Room == nil {
		return
	}

	spell := arg
	var found *Skill = ch.game.FindSkillByName(spell)
	if found == nil || found.skillType != SkillTypeSpell {
		ch.Send("You have no knowledge of that spell, try another.\r\n")
		return
	}

	prof, ok := ch.skills[found.id]
	if !ok {
		ch.Send("You have no knowledge of that spell, try another.\r\n")
		return
	}

	if prof.Cost > ch.mana {
		ch.Send("You do not have enough mana to cast that spell.\r\n")
		return
	}

	ch.mana -= prof.Cost

	ch.casting = &CastingContext{
		casting:    found,
		arguments:  arguments,
		startedAt:  time.Now(),
		complexity: prof.Complexity,
	}

	ch.Send("{WYou start uttering the words of the spell...{x\r\n")
	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		if character != ch {
			character.Send(fmt.Sprintf("\r\n{W%s{W begins casting a spell...{x\r\n", ch.GetShortDescriptionUpper(character)))
		}
	}
}

func do_spells(ch *Character, arguments string) {
	var output strings.Builder
	var count int = 0

	output.WriteString("{WYou have knowledge of the following spells:{x\r\n")

	for id, proficiency := range ch.skills {
		if ch.game.skills[id].skillType != SkillTypeSpell {
			continue
		}

		count++

		output.WriteString(fmt.Sprintf("%-18s %3d%% ", ch.game.skills[id].name, proficiency.Proficiency))

		if count%3 == 0 {
			output.WriteString("\r\n")
		}
	}

	if count%3 != 0 {
		output.WriteString("\r\n")
	}

	ch.Send(output.String())
}
