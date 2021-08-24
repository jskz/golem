/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "strings"

type Job struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Playable    bool   `json:"playable"`
}

type Race struct {
	Id          uint   `json:"id"`
	Name        string `json:"race"`
	DisplayName string `json:"display_name"`
	Playable    bool   `json:"playable"`
}

var JobsTable map[string]*Job
var RaceTable map[string]*Race

func initJobsTable() {
	JobsTable = make(map[string]*Job)

	/* Placeholder/default class */
	JobsTable["none"] = &Job{
		Id:          0,
		Name:        "none",
		DisplayName: "Tourist",
		Playable:    false,
	}
	JobsTable["warrior"] = &Job{
		Id:          1,
		Name:        "warrior",
		DisplayName: "Warrior",
		Playable:    true,
	}
	JobsTable["thief"] = &Job{
		Id:          2,
		Name:        "thief",
		DisplayName: "Thief",
		Playable:    true,
	}
	JobsTable["mage"] = &Job{
		Id:          3,
		Name:        "mage",
		DisplayName: "Mage",
		Playable:    true,
	}
	JobsTable["cleric"] = &Job{
		Id:          4,
		Name:        "cleric",
		DisplayName: "Cleric",
		Playable:    true,
	}
}

func initRaceTable() {
	RaceTable = make(map[string]*Race)

	/* Placeholder/default class */
	RaceTable["human"] = &Race{
		Id:          0,
		Name:        "human",
		DisplayName: "Human",
		Playable:    true,
	}
	RaceTable["elf"] = &Race{
		Id:          1,
		Name:        "elf",
		DisplayName: "Elf",
		Playable:    true,
	}
	RaceTable["dwarf"] = &Race{
		Id:          2,
		Name:        "dwarf",
		DisplayName: "Dwarf",
		Playable:    true,
	}
	RaceTable["ogre"] = &Race{
		Id:          3,
		Name:        "ogre",
		DisplayName: "Ogre",
		Playable:    true,
	}
}

/* Magic method to initialize constant tables */
func init() {
	initJobsTable()
	initRaceTable()
}

/* Utility lookup methods */
func FindJobByName(name string) *Job {
	for _, job := range JobsTable {
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
