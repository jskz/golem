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

type SpellContext struct {
	casting   *Skill
	startedAt time.Time
}

func do_cast(ch *Character, arguments string) {
	ch.Send("Not yet implemented, try again soon!\r\n")
}

func do_spells(ch *Character, arguments string) {
	var output strings.Builder
	var count int = 0

	output.WriteString("You have knowledge of the following spells:\r\n")

	for id, proficiency := range ch.skills {
		if ch.game.skills[id].skillType != SkillTypeSpell {
			continue
		}

		count++

		output.WriteString(fmt.Sprintf("%-15s %3d%% ", ch.game.skills[id].name, proficiency.proficiency))

		if count%3 == 0 {
			output.WriteString("\r\n")
		}
	}

	if count%3 != 0 {
		output.WriteString("\r\n")
	}

	ch.Send(output.String())
}
