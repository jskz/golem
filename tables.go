/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

type Job struct {
	Id          uint   `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type Race struct {
	Id          uint   `json:"id"`
	Name        string `json:"race"`
	DisplayName string `json:"display_name"`
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
	}
	JobsTable["warrior"] = &Job{
		Id:          1,
		Name:        "warrior",
		DisplayName: "Warrior",
	}
}

func initRaceTable() {
	RaceTable = make(map[string]*Race)

	/* Placeholder/default class */
	RaceTable["human"] = &Race{
		Id:          0,
		Name:        "human",
		DisplayName: "Human",
	}
}

/* Magic method to initialize constant tables */
func init() {
	initJobsTable()
	initRaceTable()
}
