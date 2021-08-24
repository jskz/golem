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
	"io/ioutil"
	"log"
)

const SystemDefaultConfigPath = "/etc/golem/config.json"
const ConfigPath = "etc/config.json"

var Config *AppConfiguration

/* Structure corresponding to JSON configuration file */
type AppMySQLConfiguration struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

type AppConfiguration struct {
	Port               int                   `json:"port"`
	MySQLConfiguration AppMySQLConfiguration `json:"mysql"`
}

/* Magic method is automatically called before main */
func init() {
	/* Defaults */
	Config = &AppConfiguration{
		Port: 4000,
		MySQLConfiguration: AppMySQLConfiguration{
			Host:     "mysql",
			Port:     3306,
			User:     "username",
			Password: "password",
		},
	}

	/* Attempt read of system default JSON file */
	configBytes, err := ioutil.ReadFile(SystemDefaultConfigPath)
	if err != nil {
		log.Printf("Warning: failed to read system config file: %v.\r\n", err)
	} else {
		err = json.Unmarshal(configBytes, Config)
		if err != nil {
			panic("Malformed config file.")
		}
	}

	/* Attempt read of config JSON file */
	configBytes, err = ioutil.ReadFile(ConfigPath)
	if err != nil {
		log.Printf("Warning: failed to read local config file: %v.\r\n", err)
	} else {
		err = json.Unmarshal(configBytes, Config)
		if err != nil {
			panic("Malformed config file.")
		}
	}
}
