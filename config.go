package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

const ConfigPath = "etc/config.json"

var Config *AppConfiguration

/* Structure corresponding to JSON configuration file */
type AppConfiguration struct {
	Port int `json:"port"`
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
