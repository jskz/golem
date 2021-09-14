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
)

type CastingContext struct {
	casting        *Skill
	arguments      string
	startedAt      time.Time
	ticksRemaining int
}

func do_cast(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Cast what?\r\n")
		return
	}

	args := strings.Split(arguments, " ")

	if ch.casting != nil {
		ch.Send("You are already in the middle of casting another spell!\r\n")
		return
	}

	if ch.Room == nil {
		return
	}

	spell := args[0]
	var found *Skill = ch.game.FindSkillByName(spell)
	if found == nil || found.skillType != SkillTypeSpell {
		ch.Send("You have no knowledge of that spell, try another.\r\n")
		return
	}

	ch.casting = &CastingContext{
		casting:        found,
		arguments:      strings.Join(args[1:], " "),
		startedAt:      time.Now(),
		ticksRemaining: 2,
	}

	ch.Send("{WYou start uttering the words of the spell...{x\r\n")
	for iter := ch.Room.characters.Head; iter != nil; iter = iter.Next {
		character := iter.Value.(*Character)
		if character != ch {
			character.Send(fmt.Sprintf("\r\n{W%s{W begins casting a spell...{x\r\n", ch.getShortDescriptionUpper(character)))
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

		output.WriteString(fmt.Sprintf("%-18s %3d%% ", ch.game.skills[id].name, proficiency.proficiency))

		if count%3 == 0 {
			output.WriteString("\r\n")
		}
	}

	if count%3 != 0 {
		output.WriteString("\r\n")
	}

	ch.Send(output.String())
}
