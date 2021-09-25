/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

type Skill struct {
	id        uint
	name      string
	skillType int
	handler   *goja.Callable
}

const (
	SkillTypeSkill   = 0
	SkillTypeSpell   = 1
	SkillTypePassive = 2
)

type Proficiency struct {
	id          uint
	skillId     uint
	proficiency int
	level       int
	complexity  int
	cost        int
}

func (game *Game) RegisterSkillHandler(name string, fn goja.Callable) goja.Value {
	skill := game.FindSkillByName(name)
	if skill == nil || skill.skillType != SkillTypeSkill {
		return game.vm.ToValue(nil)
	}

	skill.handler = &fn
	return game.vm.ToValue(skill)
}

func (game *Game) FindSkillByName(name string) *Skill {
	for _, skill := range game.skills {
		if skill.name == name {
			return skill
		}
	}

	return nil
}

func do_skills(ch *Character, arguments string) {
	var output strings.Builder
	var count int = 0

	output.WriteString("{WYou have knowledge of the following skills:{x\r\n")

	for id, proficiency := range ch.skills {
		if ch.game.skills[id].skillType != SkillTypeSkill && ch.game.skills[id].skillType != SkillTypePassive {
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

func do_practice(ch *Character, arguments string) {
	var output strings.Builder
	var count int = 0

	/*
	 * TODO: implement the has-arguments path where we can practice skills
	 * TODO: group output items alphasorted by skill type, colourized
	 * TBD: will train be a separate command, or do we practice attributes, too?
	 */
	output.WriteString("{WYou have knowledge of the following skills and spells:{x\r\n")

	for id, proficiency := range ch.skills {
		count++

		output.WriteString(fmt.Sprintf("%-18s %3d%% ", ch.game.skills[id].name, proficiency.proficiency))

		if count%3 == 0 {
			output.WriteString("\r\n")
		}
	}

	output.WriteString(fmt.Sprintf("\r\nYou have %d practice sessions.\r\n", ch.practices))
	ch.Send(output.String())
}

func (game *Game) LoadSkills() error {
	game.skills = make(map[uint]*Skill)

	rows, err := game.db.Query(`
		SELECT
			skills.id,
			skills.name,
			skills.type
		FROM
			skills
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var skillType string

		skill := &Skill{}

		err := rows.Scan(&skill.id, &skill.name, &skillType)
		if err != nil {
			return err
		}

		switch skillType {
		case "skill":
			skill.skillType = SkillTypeSkill
			break

		case "spell":
			skill.skillType = SkillTypeSpell
			break

		case "passive":
			skill.skillType = SkillTypePassive
			break

		default:
			err = errors.New("skill with bad enum value scanned")
			break
		}

		game.skills[skill.id] = skill
	}

	return nil
}

func (ch *Character) LoadPlayerSkills() error {
	rows, err := ch.game.db.Query(`
		SELECT
			pc_skill_proficiency.id,
			pc_skill_proficiency.skill_id,
			pc_skill_proficiency.proficiency,

			job_skill.level,
			job_skill.complexity,
			job_skill.cost
		FROM
			pc_skill_proficiency
		INNER JOIN
			job_skill
		ON
			job_skill.id = pc_skill_proficiency.player_character_id
		WHERE
			pc_skill_proficiency.player_character_id = ?
	`, ch.Id)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		proficiency := &Proficiency{}

		err := rows.Scan(&proficiency.id, &proficiency.skillId, &proficiency.proficiency, &proficiency.level, &proficiency.complexity, &proficiency.cost)
		if err != nil {
			return err
		}

		ch.skills[proficiency.id] = proficiency
	}

	return nil
}
