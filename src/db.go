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

var JobTable map[uint]*Job
var RaceTable map[uint]*Race

func (game *Game) LoadRaceTable() {
	log.Printf("Loading races.\r\n")

	RaceTable = make(map[uint]*Race)

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

		RaceTable[race.Id] = race
	}

	log.Printf("Loaded %d races from database.\r\n", len(RaceTable))
}

func (game *Game) LoadJobTable() {
	log.Printf("Loading jobs.\r\n")

	JobTable = make(map[uint]*Job)

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			display_name,
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

		err := rows.Scan(&job.Id, &job.Name, &job.DisplayName, &job.Playable)
		if err != nil {
			log.Printf("Unable to scan job row: %v.\r\n", err)
			continue
		}

		JobTable[job.Id] = job
	}

	log.Printf("Loaded %d jobs from database.\r\n", len(JobTable))
}

/* Utility lookup methods */
func FindJobByName(name string) *Job {
	for _, job := range JobTable {
		if strings.Compare(name, job.Name) == 0 {
			return job
		}
	}

	return nil
}

func FindRaceByName(name string) *Race {
	for _, race := range RaceTable {
		if strings.Compare(name, race.Name) == 0 {
			return race
		}
	}

	return nil
}
