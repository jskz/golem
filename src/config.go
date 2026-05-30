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

/* Structure corresponding to JSON configuration file */
type AppMySQLConfiguration struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type AppDatabaseConfiguration struct {
	Driver   string `json:"driver"`
	DSN      string `json:"dsn"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Path     string `json:"path"`
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
	MySQLConfiguration     AppMySQLConfiguration     `json:"mysql"`
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
		MySQLConfiguration:    defaultMySQLConfiguration(),
	}

	/* Attempt read of config JSON file */
	configBytes, err := os.ReadFile("etc/config.json")
	if err != nil {
		log.Printf("Warning: failed to read local config file: %v.\r\n", err)
		Config.normalizeDatabaseConfiguration(nil)
	} else {
		var configFields map[string]json.RawMessage
		err = json.Unmarshal(configBytes, &configFields)
		if err != nil {
			panic("Malformed config file.")
		}

		err = json.Unmarshal(configBytes, Config)
		if err != nil {
			panic("Malformed config file.")
		}

		Config.normalizeDatabaseConfiguration(configFields)
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

func defaultMySQLConfiguration() AppMySQLConfiguration {
	return AppMySQLConfiguration{
		Host:     "mysql",
		Port:     3306,
		User:     "username",
		Password: "password",
		Database: "database",
	}
}

func defaultDatabaseConfiguration() AppDatabaseConfiguration {
	mysqlConfig := defaultMySQLConfiguration()

	return AppDatabaseConfiguration{
		Driver:   databaseDriverMySQL,
		Host:     mysqlConfig.Host,
		Port:     mysqlConfig.Port,
		User:     mysqlConfig.User,
		Password: mysqlConfig.Password,
		Database: mysqlConfig.Database,
		Path:     "etc/golem.sqlite3",
	}
}

func databaseConfigurationFromMySQL(mysqlConfig AppMySQLConfiguration) AppDatabaseConfiguration {
	return AppDatabaseConfiguration{
		Driver:   databaseDriverMySQL,
		Host:     mysqlConfig.Host,
		Port:     mysqlConfig.Port,
		User:     mysqlConfig.User,
		Password: mysqlConfig.Password,
		Database: mysqlConfig.Database,
		Path:     "etc/golem.sqlite3",
	}
}

func (config *AppConfiguration) normalizeDatabaseConfiguration(fields map[string]json.RawMessage) {
	if fields != nil {
		_, hasDatabaseConfig := fields["database"]
		_, hasLegacyMySQLConfig := fields["mysql"]

		if !hasDatabaseConfig && hasLegacyMySQLConfig {
			config.DatabaseConfiguration = databaseConfigurationFromMySQL(config.MySQLConfiguration)
		}
	}

	if config.DatabaseConfiguration.Driver == "" {
		config.DatabaseConfiguration.Driver = databaseDriverMySQL
	}

	if config.DatabaseConfiguration.Path == "" {
		config.DatabaseConfiguration.Path = "etc/golem.sqlite3"
	}
}
