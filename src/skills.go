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
	"log"
	"sort"
	"strings"

	"github.com/dop251/goja"
)

type Skill struct {
	Id        uint
	Name      string
	SkillType int
	Intent    string
	Handler   *goja.Callable
}

const (
	SkillTypeSkill   = 0
	SkillTypeSpell   = 1
	SkillTypePassive = 2
)

const (
	SkillIntentNone      = "none"
	SkillIntentOffensive = "offensive"
	SkillIntentCurative  = "curative"
)

type JobSkill struct {
	Id int `json:"id"`

	Job   *Job   `json:"job"`
	Skill *Skill `json:"skill"`

	Level      int `json:"level"`
	Complexity int `json:"complexity"`
	Cost       int `json:"cost"`
}

type Proficiency struct {
	Job *Job `json:"job"`

	Id          uint `json:"id"`
	SkillId     uint `json:"skillId"`
	Proficiency int  `json:"proficiency"`
	Level       int  `json:"level"`
	Complexity  int  `json:"complexity"`
	Cost        int  `json:"cost"`
}

var SkillIntentColourTable map[string]string

func init() {
	SkillIntentColourTable = make(map[string]string)
	SkillIntentColourTable[SkillIntentNone] = "{x"
	SkillIntentColourTable[SkillIntentOffensive] = "{R"
	SkillIntentColourTable[SkillIntentCurative] = "{C"
}

func (game *Game) RegisterSkillHandler(name string, fn goja.Callable) goja.Value {
	skill := game.FindSkillByName(name)
	if skill == nil || skill.SkillType != SkillTypeSkill {
		return game.vm.ToValue(nil)
	}

	skill.Handler = &fn
	return game.vm.ToValue(skill)
}

func (game *Game) FindSkillByID(id uint) *Skill {
	for _, skill := range game.skills {
		if skill.Id == id {
			return skill
		}
	}

	return nil
}

func (game *Game) FindSkillByName(name string) *Skill {
	for _, skill := range game.skills {
		if skill.Name == name {
			return skill
		}
	}

	return nil
}

func (ch *Character) FindProficiencyByName(name string) *Proficiency {
	for _, skill := range ch.Skills {
		if ch.Game.skills[skill.SkillId].Name == name {
			return ch.Skills[skill.SkillId]
		}
	}

	return nil
}

func (ch *Character) syncJobSkills() error {
	for iter := ch.Job.Skills.Head; iter != nil; iter = iter.Next {
		jobSkill := iter.Value.(*JobSkill)

		if uint(jobSkill.Level) > ch.Level {
			continue
		}

		_, ok := ch.Skills[jobSkill.Skill.Id]
		if !ok {
			proficiency := &Proficiency{}

			proficiency.SkillId = jobSkill.Skill.Id
			proficiency.Complexity = jobSkill.Complexity
			proficiency.Level = jobSkill.Level
			proficiency.Cost = jobSkill.Cost
			proficiency.Job = jobSkill.Job
			proficiency.Proficiency = 0

			/* Try to create the pc_skill_proficiency relationship before finalizing this skill attach */
			res, err := ch.Game.db.Exec(`
			INSERT INTO
				pc_skill_proficiency(player_character_id, skill_id, job_id, proficiency)
			VALUES
				(?, ?, ?, ?)
			`, ch.Id, jobSkill.Skill.Id, jobSkill.Job.Id, 0)
			if err != nil {
				return err
			}

			var insertId int64

			insertId, err = res.LastInsertId()
			if err != nil {
				return err
			}

			/* We have successfully insert the PC proficiency, attach it in-memory and continue */
			proficiency.Id = uint(insertId)
			ch.Skills[jobSkill.Skill.Id] = proficiency
		}
	}

	return nil
}

func do_skills(ch *Character, arguments string) {
	var output strings.Builder
	var count int = 0

	output.WriteString("{WYou have knowledge of the following skills:{x\r\n")

	for id, proficiency := range ch.Skills {
		if ch.Game.skills[id].SkillType != SkillTypeSkill && ch.Game.skills[id].SkillType != SkillTypePassive {
			continue
		}

		count++

		output.WriteString(fmt.Sprintf("%s%-18s %3d%% ", SkillIntentColourTable[ch.Game.skills[id].Intent], ch.Game.skills[id].Name, proficiency.Proficiency))

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
	var firstArgument string = ""
	var output strings.Builder
	var count int = 0

	firstArgument, _ = oneArgument(arguments)

	if firstArgument != "" {
		var trainerFound bool = false

		if ch.Room == nil {
			ch.Send("You can't practice here.\r\n")
			return
		}

		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch.Flags&CHAR_PRACTICE != 0 {
				trainerFound = true
			}
		}

		if !trainerFound {
			ch.Send("There is nobody here who can teach you.\r\n")
			return
		}

		skill := ch.Game.FindSkillByName(firstArgument)
		if skill == nil {
			ch.Send("You can't practice that.\r\n")
			return
		}

		prof, ok := ch.Skills[skill.Id]
		if !ok {
			ch.Send("You can't practice that.\r\n")
			return
		}

		if ch.Practices < prof.Complexity {
			ch.Send("You don't have enough practice sessions.\r\n")
			return
		}

		ch.Practices -= prof.Complexity
		prof.Proficiency++
		ch.Send(fmt.Sprintf("{WYou practice %s!{x\r\n", skill.Name))
		return
	}

	output.WriteString("{WYou have knowledge of the following skills and spells:{x\r\n")

	var skills []string = []string{}
	var proficiencies map[string]int = make(map[string]int)

	for _, proficiency := range ch.Skills {
		found := false

		_, ok := ch.Game.skills[proficiency.SkillId]
		if !ok {
			log.Printf("Player had a proficiency with a nonexistent id %d.\r\n", proficiency.SkillId)
			continue
		}

		for _, c := range skills {
			if c == ch.Game.skills[proficiency.SkillId].Name {
				found = true
			}
		}

		if !found {
			var skillName string = fmt.Sprintf("%s%s{x", SkillIntentColourTable[ch.Game.skills[proficiency.SkillId].Intent], ch.Game.skills[proficiency.SkillId].Name)

			if strings.ContainsRune(skillName, ' ') && ch.Game.skills[proficiency.SkillId].SkillType == SkillTypeSpell {
				skillName = fmt.Sprintf("'%s'", skillName)
			}

			skills = append(skills, skillName)
			proficiencies[skillName] = proficiency.Proficiency
		}
	}

	sort.Strings(skills)

	for _, proficiency := range skills {
		count++

		output.WriteString(fmt.Sprintf("%-18s %3d%% ", proficiency, proficiencies[proficiency]))

		if count%3 == 0 {
			output.WriteString("\r\n")
		}
	}

	output.WriteString(fmt.Sprintf("\r\n{xYou have %d practice sessions.\r\n", ch.Practices))
	ch.Send(output.String())
}

func (ch *Character) SaveSkills() error {
	return nil
}

func (game *Game) LoadSkills() error {
	game.skills = make(map[uint]*Skill)

	rows, err := game.db.Query(`
		SELECT
			skills.id,
			skills.name,
			skills.type,
			skills.intent
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

		err := rows.Scan(&skill.Id, &skill.Name, &skillType, &skill.Intent)
		if err != nil {
			return err
		}

		switch skillType {
		case "skill":
			skill.SkillType = SkillTypeSkill

		case "spell":
			skill.SkillType = SkillTypeSpell

		case "passive":
			skill.SkillType = SkillTypePassive

		default:
			return errors.New("skill with bad enum value scanned")
		}

		game.skills[skill.Id] = skill
	}

	return nil
}

func (ch *Character) LoadPlayerSkills() error {
	rows, err := ch.Game.db.Query(`
		SELECT
			pc_skill_proficiency.id,
			pc_skill_proficiency.skill_id,
			pc_skill_proficiency.job_id,
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

		var jobId uint = 0

		err := rows.Scan(&proficiency.Id, &proficiency.SkillId, &jobId, &proficiency.Proficiency, &proficiency.Level, &proficiency.Complexity, &proficiency.Cost)
		if err != nil {
			return err
		}

		for iter := Jobs.Head; iter != nil; iter = iter.Next {
			job := iter.Value.(*Job)

			if job.Id == jobId {
				proficiency.Job = job
			}
		}

		if proficiency.Job == nil {
			log.Printf("Failed to attach PC proficiency because its job ID did not exist.\r\n")
			return nil
		}

		ch.Skills[proficiency.SkillId] = proficiency
	}

	return nil
}
