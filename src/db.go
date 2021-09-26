/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
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

		err := rows.Scan(&terrain.id, &terrain.name, &terrain.mapGlyph, &terrain.movementCost, &terrain.flags)
		if err != nil {
			log.Printf("Unable to scan terrain: %v.\r\n", err)
			continue
		}

		TerrainTable[terrain.id] = terrain
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

		Jobs.Insert(job)
	}

	log.Printf("Loaded %d jobs from database.\r\n", Jobs.Count)
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
