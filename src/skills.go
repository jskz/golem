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
var trainableStats []int = []int{
	STAT_STRENGTH,
	STAT_DEXTERITY,
	STAT_INTELLIGENCE,
	STAT_WISDOM,
	STAT_CONSTITUTION,
	STAT_CHARISMA,
	STAT_LUCK,
}

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
	for jobSkill := range ch.Job.Skills.All() {
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

func do_train(ch *Character, arguments string) {
	firstArgument, _ := OneArgument(arguments)

	if ch.Flags&CHAR_IS_PLAYER == 0 {
		return
	}

	if !characterHasTrainer(ch, CHAR_TRAIN) {
		ch.Send("You can't do that here.\r\n")
		return
	}

	if firstArgument == "" {
		ch.Send(fmt.Sprintf("You have %d training sessions.\r\n%s", ch.Trains, trainOptions(ch)))
		return
	}

	switch firstArgument {
	case "hp", "health":
		trainResource(ch, "health")
		return
	case "mana":
		trainResource(ch, "mana")
		return
	case "move", "stamina":
		trainResource(ch, "stamina")
		return
	}

	stat := trainStatByName(firstArgument)
	if stat == STAT_NONE || stat >= len(ch.Stats) {
		ch.Send(trainOptions(ch))
		return
	}

	statName := StatName(stat)
	if ch.Stats[stat] >= trainStatMaximum(ch, stat) {
		ch.Send(fmt.Sprintf("Your %s is already at maximum.\r\n", statName))
		return
	}

	if ch.Trains < 1 {
		ch.Send("You don't have enough training sessions.\r\n")
		return
	}

	ch.Trains--
	ch.Stats[stat]++
	ch.Send(fmt.Sprintf("{WYour %s increases!{x\r\n", statName))
	for rch := range ch.Room.Characters.All() {
		if rch != ch {
			rch.Send(fmt.Sprintf("\r\n{W%s{W's %s increases!{x\r\n", ch.GetShortDescriptionUpper(rch), statName))
		}
	}
}

func characterHasTrainer(ch *Character, flag int) bool {
	if ch == nil || ch.Room == nil || ch.Room.Characters == nil {
		return false
	}

	for rch := range ch.Room.Characters.All() {
		if rch != ch && rch.Flags&CHAR_IS_PLAYER == 0 && rch.Flags&flag != 0 {
			return true
		}
	}

	return false
}

func trainOptions(ch *Character) string {
	var output strings.Builder

	output.WriteString("You can train:")
	for _, stat := range trainableStats {
		if stat >= len(ch.Stats) || ch.Stats[stat] >= trainStatMaximum(ch, stat) {
			continue
		}

		output.WriteString(fmt.Sprintf(" %s", StatName(stat)))
	}

	output.WriteString(" health mana stamina.\r\n")
	return output.String()
}

func trainStatByName(name string) int {
	switch name {
	case "str", "strength":
		return STAT_STRENGTH
	case "dex", "dexterity":
		return STAT_DEXTERITY
	case "int", "intelligence":
		return STAT_INTELLIGENCE
	case "wis", "wisdom":
		return STAT_WISDOM
	case "con", "constitution":
		return STAT_CONSTITUTION
	case "cha", "charisma":
		return STAT_CHARISMA
	case "lck", "luck":
		return STAT_LUCK
	default:
		return STAT_NONE
	}
}

func trainStatMaximum(ch *Character, stat int) int {
	maximum := 20

	if ch.Job != nil && ch.Job.PrimaryAttribute == stat {
		maximum += 2
	}

	if ch.Race != nil && ch.Race.PrimaryAttribute == stat {
		maximum += 2
	}

	return maximum
}

func trainResource(ch *Character, resource string) {
	if ch.Trains < 1 {
		ch.Send("You don't have enough training sessions.\r\n")
		return
	}

	ch.Trains--

	switch resource {
	case "health":
		ch.MaxHealth += 10
		ch.Health += 10
	case "mana":
		ch.MaxMana += 10
		ch.Mana += 10
	case "stamina":
		ch.MaxStamina += 10
		ch.Stamina += 10
	}

	ch.Send(fmt.Sprintf("{WYour %s increases!{x\r\n", resource))
	for rch := range ch.Room.Characters.All() {
		if rch != ch {
			rch.Send(fmt.Sprintf("\r\n{W%s{W's %s increases!{x\r\n", ch.GetShortDescriptionUpper(rch), resource))
		}
	}
}

func do_practice(ch *Character, arguments string) {
	var firstArgument string = ""
	var output strings.Builder
	var count int = 0

	firstArgument, _ = OneArgument(arguments)

	if firstArgument != "" {
		var trainerFound bool = false

		if ch.Room == nil {
			ch.Send("You can't practice here.\r\n")
			return
		}

		for rch := range ch.Room.Characters.All() {
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

		if prof.Proficiency >= 100 {
			ch.Send("You have already mastered this proficiency.\r\n")
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

	return rows.Err()
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
			job_skill.job_id = pc_skill_proficiency.job_id
			AND job_skill.skill_id = pc_skill_proficiency.skill_id
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

		for job := range Jobs.All() {
			if job.Id == jobId {
				proficiency.Job = job
			}
		}

		if proficiency.Job == nil {
			return fmt.Errorf("failed to attach PC proficiency %d for player %d: job ID %d did not exist", proficiency.Id, ch.Id, jobId)
		}

		ch.Skills[proficiency.SkillId] = proficiency
	}

	return rows.Err()
}
