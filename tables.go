/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

type Job struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

var JobsTable map[string]*Job

/* Magic method to initialize the job table */
func init() {
	JobsTable = make(map[string]*Job)

	/* Placeholder/default class */
	JobsTable["none"] = &Job{
		Name:        "none",
		DisplayName: "Tourist",
	}
	JobsTable["warrior"] = &Job{
		Name:        "warrior",
		DisplayName: "Warrior",
	}
}
