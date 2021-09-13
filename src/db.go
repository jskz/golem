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

var Jobs *LinkedList
var Races *LinkedList

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

	log.Printf("Loaded %d races from database.\r\n", Races.count)
}

func (game *Game) LoadJobTable() {
	log.Printf("Loading jobs.\r\n")

	Jobs = NewLinkedList()

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

		Jobs.Insert(job)
	}

	log.Printf("Loaded %d jobs from database.\r\n", Jobs.count)
}

/* Utility lookup methods */
func FindJobByName(name string) *Job {
	for iter := Jobs.head; iter != nil; iter = iter.next {
		job := iter.value.(*Job)

		if strings.Compare(name, job.Name) == 0 {
			return job
		}
	}

	return nil
}

func FindRaceByName(name string) *Race {
	for iter := Races.head; iter != nil; iter = iter.next {
		race := iter.value.(*Race)

		if strings.Compare(name, race.Name) == 0 {
			return race
		}
	}

	return nil
}

func FindJobByID(id uint) *Job {
	for iter := Jobs.head; iter != nil; iter = iter.next {
		job := iter.value.(*Job)

		if job.Id == id {
			return job
		}
	}

	return nil
}

func FindRaceByID(id uint) *Race {
	for iter := Races.head; iter != nil; iter = iter.next {
		race := iter.value.(*Race)

		if race.Id == id {
			return race
		}
	}

	return nil
}
