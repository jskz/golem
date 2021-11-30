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
var Jobs *LinkedList
var Races *LinkedList

func (game *Game) LoadTerrain() {
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
		return
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

	log.Printf("Loaded %d terrain types from database.\r\n", len(TerrainTable))
}

func (game *Game) LoadRaceTable() {
	log.Printf("Loading races.\r\n")

	Races = NewLinkedList()

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			display_name,
			playable
		FROM
			races
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		race := &Race{}

		err := rows.Scan(&race.Id, &race.Name, &race.DisplayName, &race.Playable)
		if err != nil {
			log.Printf("Unable to scan race row: %v.\r\n", err)
			continue
		}

		Races.Insert(race)
	}

	log.Printf("Loaded %d races from database.\r\n", Races.Count)
}

func (game *Game) LoadJobTable() {
	log.Printf("Loading jobs.\r\n")

	Jobs = NewLinkedList()

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			display_name,
			experience_required_modifier,
			playable
		FROM
			jobs
		WHERE
			deleted_at IS NULL
	`)
	if err != nil {
		return
	}

	defer rows.Close()

	for rows.Next() {
		job := &Job{}

		err := rows.Scan(&job.Id, &job.Name, &job.DisplayName, &job.ExperienceRequiredModifier, &job.Playable)
		if err != nil {
			log.Printf("Unable to scan job row: %v.\r\n", err)
			continue
		}

		job.Skills = NewLinkedList()
		Jobs.Insert(job)
	}

	log.Printf("Loaded %d jobs from database.\r\n", Jobs.Count)
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

	log.Printf("Loaded %d job-skill relations from database.\r\n", results)
	return nil
}

/* Utility lookup methods */
func FindJobByName(name string) *Job {
	for iter := Jobs.Head; iter != nil; iter = iter.Next {
		job := iter.Value.(*Job)

		if strings.Compare(name, job.Name) == 0 {
			return job
		}
	}

	return nil
}

func FindRaceByName(name string) *Race {
	for iter := Races.Head; iter != nil; iter = iter.Next {
		race := iter.Value.(*Race)

		if strings.Compare(name, race.Name) == 0 {
			return race
		}
	}

	return nil
}

func FindJobByID(id uint) *Job {
	for iter := Jobs.Head; iter != nil; iter = iter.Next {
		job := iter.Value.(*Job)

		if job.Id == id {
			return job
		}
	}

	return nil
}

func FindRaceByID(id uint) *Race {
	for iter := Races.Head; iter != nil; iter = iter.Next {
		race := iter.Value.(*Race)

		if race.Id == id {
			return race
		}
	}

	return nil
}
