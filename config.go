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
	}

	/* Attempt read of config JSON file */
	configBytes, err := ioutil.ReadFile(ConfigPath)
	if err != nil {
		log.Printf("Warning: failed to read config file: %v.\r\n", err)
		return
	}

	err = json.Unmarshal(configBytes, Config)
	if err != nil {
		panic("Malformed config file.")
	}
}
