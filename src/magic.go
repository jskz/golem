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
	Casting     *Skill    `json:"casting"`
	Arguments   string    `json:"arguments"`
	StartedAt   time.Time `json:"startedAt"`
	Complexity  int       `json:"complexity"`
	Proficiency int       `json:"ability"`
}

func (ch *Character) onCastingUpdate() {
	if int(time.Since(ch.Casting.StartedAt).Seconds()) > ch.Casting.Complexity {
		ch.Send("\r\n{WYou finish casting the magic spell.{x\r\n")

		if ch.Room != nil {
			for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
				rch := iter.Value.(*Character)

				if !rch.IsEqual(ch) {
					rch.Send(fmt.Sprintf("\r\n{W%s{W finishes casting their magic spell.{W{x\r\n", ch.GetShortDescriptionUpper(rch)))
				}
			}
		}

		if ch.Casting.Casting.Handler != nil {
			fn := *ch.Casting.Casting.Handler

			fn(ch.Game.vm.ToValue(ch.Casting), ch.Game.vm.ToValue(ch), ch.Game.vm.ToValue(ch.Casting.Arguments))
		}

		ch.Casting = nil
	}
}

func (game *Game) RegisterSpellHandler(name string, fn goja.Callable) goja.Value {
	spell := game.FindSkillByName(name)
	if spell == nil || spell.SkillType != SkillTypeSpell {
		return game.vm.ToValue(nil)
	}

	spell.Handler = &fn
	return game.vm.ToValue(spell)
}

func do_cast(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Cast what?\r\n")
		return
	}

	arg, arguments := oneArgument(arguments)

	if ch.Casting != nil {
		ch.Send("You are already in the middle of casting another spell!\r\n")
		return
	}

	if ch.Room == nil {
		return
	}

	spell := arg
	var found *Skill = ch.Game.FindSkillByName(spell)
	if found == nil || found.SkillType != SkillTypeSpell {
		ch.Send("You have no knowledge of that spell, try another.\r\n")
		return
	}

	prof, ok := ch.Skills[found.Id]
	if !ok {
		ch.Send("You have no knowledge of that spell, try another.\r\n")
		return
	}

	if found.Intent == SkillIntentOffensive && ch.Room.Flags&ROOM_SAFE != 0 {
		ch.Send("You can't cast that here.\r\n")
		return
	}

	if prof.Cost > ch.Mana {
		ch.Send("You do not have enough mana to cast that spell.\r\n")
		return
	}

	ch.Mana -= prof.Cost

	ch.Casting = &CastingContext{
		Casting:     found,
		Arguments:   arguments,
		StartedAt:   time.Now(),
		Complexity:  prof.Complexity,
		Proficiency: prof.Proficiency,
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

	for id, proficiency := range ch.Skills {
		if ch.Game.skills[id].SkillType != SkillTypeSpell {
			continue
		}

		count++

		var skillName string = fmt.Sprintf("%s%s{x", SkillIntentColourTable[ch.Game.skills[id].Intent], ch.Game.skills[id].Name)
		if strings.ContainsRune(skillName, ' ') {
			skillName = fmt.Sprintf("'%s'", skillName)
		}

		output.WriteString(fmt.Sprintf("%-18s %3d%% ", skillName, proficiency.Proficiency))

		if count%3 == 0 {
			output.WriteString("\r\n")
		}
	}

	if count%3 != 0 {
		output.WriteString("\r\n")
	}

	ch.Send(output.String())
}
