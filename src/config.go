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
	"time"

	"github.com/getsentry/sentry-go"
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

type AppRedisConfiguration struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
}

type AppProfilingConfiguration struct {
	Enabled bool `json:"enabled"`
	Port    int  `json:"port"`
}

type AppSentryConfiguration struct {
	DSN     string `json:"dsn"`
	Enabled bool   `json:"enabled"`
}

type AppWebConfiguration struct {
	PublicRoot string `json:"publicRoot"`
}

type AppConfiguration struct {
	Port                   int                       `json:"port"`
	MySQLConfiguration     AppMySQLConfiguration     `json:"mysql"`
	RedisConfiguration     AppRedisConfiguration     `json:"redis"`
	SentryConfiguration    AppSentryConfiguration    `json:"sentry"`
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
		Port: 4000,
		MySQLConfiguration: AppMySQLConfiguration{
			Host:     "mysql",
			Port:     3306,
			User:     "username",
			Password: "password",
		},
	}

	/* Attempt read of config JSON file */
	configBytes, err := ioutil.ReadFile("etc/config.json")
	if err != nil {
		log.Printf("Warning: failed to read local config file: %v.\r\n", err)
	} else {
		err = json.Unmarshal(configBytes, Config)
		if err != nil {
			panic("Malformed config file.")
		}
	}

	if Config.SentryConfiguration.Enabled {
		err := sentry.Init(sentry.ClientOptions{
			Dsn: Config.SentryConfiguration.DSN,
		})

		if err != nil {
			log.Panicf("failed to initialize Sentry: %v.\r\n", err)
		}

		log.Printf("Enabled Sentry error reporting.\r\n")
		defer sentry.Flush(5 * time.Second)
	}

	/* Read greeting */
	Config.greeting, err = ioutil.ReadFile("etc/GREETING.ANS")
	if err != nil {
		log.Printf("Warning: failed to read greeting ANSI file: %v.\r\n", err)
		Config.greeting = []byte(string(""))
	}

	/* Read MOTD */
	Config.motd, err = ioutil.ReadFile("etc/MOTD.ANS")
	if err != nil {
		log.Printf("Warning: failed to read MOTD ANSI file: %v.\r\n", err)
		Config.motd = []byte(string(""))
	}

	/* Read death ANSI */
	Config.death, err = ioutil.ReadFile("etc/DEATH.ANS")
	if err != nil {
		log.Printf("Warning: failed to read death ANSI file: %v.\r\n", err)
		Config.death = []byte(string(""))
	}
}
