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
	"log"
	"strings"
)

var TerrainTable map[int]*Terrain
var Jobs *LinkedList[*Job]
var Races *LinkedList[*Race]

func (game *Game) LoadTerrain() error {
	log.Printf("Loading terrain types.\r\n")

	TerrainTable = make(map[int]*Terrain)

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			glyph_colour,
			map_glyph,
			movement_cost,
			flags
		FROM
			terrain
	`)
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		terrain := &Terrain{}

		err := rows.Scan(&terrain.Id, &terrain.Name, &terrain.GlyphColour, &terrain.MapGlyph, &terrain.MovementCost, &terrain.Flags)
		if err != nil {
			log.Printf("Unable to scan terrain: %v.\r\n", err)
			continue
		}

		TerrainTable[terrain.Id] = terrain
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d terrain types from database.\r\n", len(TerrainTable))
	return nil
}

func (game *Game) LoadRaceTable() error {
	log.Printf("Loading races.\r\n")

	Races = NewLinkedList[*Race]()

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			display_name,
			playable,
			primary_attribute
		FROM
			races
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		race := &Race{}
		var primaryAttribute string = "none"

		err := rows.Scan(&race.Id, &race.Name, &race.DisplayName, &race.Playable, &primaryAttribute)
		if err != nil {
			log.Printf("Unable to scan race row: %v.\r\n", err)
			continue
		}

		race.PrimaryAttribute = FindStatByName(primaryAttribute)
		Races.Insert(race)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d races from database.\r\n", Races.Count)
	return nil
}

func (game *Game) LoadJobTable() error {
	log.Printf("Loading jobs.\r\n")

	Jobs = NewLinkedList[*Job]()

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			display_name,
			experience_required_modifier,
			playable,
			primary_attribute,
			health_gain_min,
			health_gain_max,
			mana_gain_divisor,
			stamina_gain_min,
			stamina_gain_max,
			stamina_gain_floor
		FROM
			jobs
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		job := &Job{}
		var primaryAttribute string = "none"

		err := rows.Scan(
			&job.Id,
			&job.Name,
			&job.DisplayName,
			&job.ExperienceRequiredModifier,
			&job.Playable,
			&primaryAttribute,
			&job.HealthGainMin,
			&job.HealthGainMax,
			&job.ManaGainDivisor,
			&job.StaminaGainMin,
			&job.StaminaGainMax,
			&job.StaminaGainFloor,
		)
		if err != nil {
			log.Printf("Unable to scan job row: %v.\r\n", err)
			continue
		}

		job.Skills = NewLinkedList[*JobSkill]()
		job.PrimaryAttribute = FindStatByName(primaryAttribute)
		Jobs.Insert(job)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d jobs from database.\r\n", Jobs.Count)
	return nil
}

/* Load job-skill relationships */
func (game *Game) LoadJobSkills() error {
	log.Printf("Loading job-skill relationships.\r\n")

	rows, err := game.db.Query(`
		SELECT
			id,
			job_id,
			skill_id,
			level,
			complexity,
			cost
		FROM
			job_skill
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	var results int = 0

	for rows.Next() {
		jobSkill := &JobSkill{}

		var jobId uint = 0
		var skillId uint = 0

		err := rows.Scan(&jobSkill.Id, &jobId, &skillId, &jobSkill.Level, &jobSkill.Complexity, &jobSkill.Cost)
		if err != nil {
			log.Printf("Unable to scan job row: %v.\r\n", err)
			continue
		}

		jobSkill.Job = FindJobByID(jobId)
		if jobSkill.Job == nil {
			return errors.New("failed to attach job during job-skill relations load")
		}

		jobSkill.Skill = game.FindSkillByID(skillId)
		if jobSkill.Skill == nil {
			return errors.New("failed to attach skill during job-skill relations load")
		}

		jobSkill.Job.Skills.Insert(jobSkill)
		results++
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d job-skill relations from database.\r\n", results)
	return nil
}

/* Utility lookup methods */
func FindJobByName(name string) *Job {
	for job := range Jobs.All() {
		if strings.Compare(name, job.Name) == 0 {
			return job
		}
	}

	return nil
}

func FindRaceByName(name string) *Race {
	for race := range Races.All() {
		if strings.Compare(name, race.Name) == 0 {
			return race
		}
	}

	return nil
}

func FindJobByID(id uint) *Job {
	for job := range Jobs.All() {
		if job.Id == id {
			return job
		}
	}

	return nil
}

func FindRaceByID(id uint) *Race {
	for race := range Races.All() {
		if race.Id == id {
			return race
		}
	}

	return nil
}
