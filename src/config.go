/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"encoding/json"
	"log"
	"os"
)

var Config *AppConfiguration

const defaultDatabasePath = "etc/golem.sqlite3"

/* Structure corresponding to JSON configuration file */
type AppDatabaseConfiguration struct {
	Driver string `json:"driver"`
	DSN    string `json:"dsn"`
	Path   string `json:"path"`
}

type AppProfilingConfiguration struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type AppWebConfiguration struct {
	PublicRoot string `json:"publicRoot"`
}

type AppConfiguration struct {
	HashSalt               string                    `json:"hashSalt"`
	Port                   int                       `json:"port"`
	DatabaseConfiguration  AppDatabaseConfiguration  `json:"database"`
	ProfilingConfiguration AppProfilingConfiguration `json:"profiling"`
	WebConfiguration       AppWebConfiguration       `json:"web"`

	greeting []byte
	motd     []byte
	death    []byte
}

/* Magic method is automatically called before main */
func init() {
	/* Defaults */
	Config = &AppConfiguration{
		Port:                  4000,
		DatabaseConfiguration: defaultDatabaseConfiguration(),
	}

	/* Attempt read of config JSON file */
	configBytes, err := os.ReadFile("etc/config.json")
	if err != nil {
		log.Printf("Warning: failed to read local config file: %v.\r\n", err)
		Config.normalizeDatabaseConfiguration()
	} else {
		err = json.Unmarshal(configBytes, Config)
		if err != nil {
			panic("Malformed config file.")
		}

		Config.normalizeDatabaseConfiguration()
	}

	/* Read greeting */
	Config.greeting, err = os.ReadFile("etc/GREETING.ANS")
	if err != nil {
		log.Printf("Warning: failed to read greeting ANSI file: %v.\r\n", err)
		Config.greeting = []byte(string(""))
	}

	/* Read MOTD */
	Config.motd, err = os.ReadFile("etc/MOTD.ANS")
	if err != nil {
		log.Printf("Warning: failed to read MOTD ANSI file: %v.\r\n", err)
		Config.motd = []byte(string(""))
	}

	/* Read death ANSI */
	Config.death, err = os.ReadFile("etc/DEATH.ANS")
	if err != nil {
		log.Printf("Warning: failed to read death ANSI file: %v.\r\n", err)
		Config.death = []byte(string(""))
	}
}

func defaultDatabaseConfiguration() AppDatabaseConfiguration {
	return AppDatabaseConfiguration{
		Driver: databaseDriverSQLite,
		Path:   defaultDatabasePath,
	}
}

func (config *AppConfiguration) normalizeDatabaseConfiguration() {
	if config.DatabaseConfiguration.Driver == "" {
		config.DatabaseConfiguration.Driver = databaseDriverSQLite
	}

	if config.DatabaseConfiguration.Path == "" {
		config.DatabaseConfiguration.Path = defaultDatabasePath
	}
}
