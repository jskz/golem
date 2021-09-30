/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dop251/goja"

	_ "github.com/go-sql-driver/mysql"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
)

type Game struct {
	db *sql.DB
	vm *goja.Runtime

	eventHandlers map[string]*LinkedList

	Characters *LinkedList `json:"characters"`
	Fights     *LinkedList `json:"fights"`
	Zones      *LinkedList `json:"zones"`

	clients map[*Client]bool
	skills  map[uint]*Skill
	world   map[uint]*Room

	register        chan *Client
	unregister      chan *Client
	quitRequest     chan *Client
	shutdownRequest chan bool
	clientMessage   chan ClientTextMessage
}

func NewGame() (*Game, error) {
	var err error

	/* Create the game world instance and initialize variables & channels */
	game := &Game{}

	game.clients = make(map[*Client]bool)
	game.register = make(chan *Client)
	game.unregister = make(chan *Client)
	game.quitRequest = make(chan *Client)
	game.shutdownRequest = make(chan bool)
	game.clientMessage = make(chan ClientTextMessage)

	game.Characters = NewLinkedList()
	game.Fights = NewLinkedList()

	/* Initialize services we'll inject elsewhere through the game instance. */
	game.db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?multiStatements=true&parseTime=true",
		Config.MySQLConfiguration.User,
		Config.MySQLConfiguration.Password,
		Config.MySQLConfiguration.Host,
		Config.MySQLConfiguration.Port,
		Config.MySQLConfiguration.Database))
	if err != nil {
		return nil, err
	}

	/* Validate we can interact with the DSN */
	err = game.db.Ping()
	if err != nil {
		return nil, err
	}

	game.db.SetConnMaxLifetime(time.Second * 30)
	game.db.SetMaxOpenConns(10)
	game.db.SetMaxIdleConns(10)

	/* Attempt new migrations at startup */
	driver, _ := mysql.WithInstance(game.db, &mysql.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file:///migrations",
		"mysql",
		driver,
	)
	if err != nil {
		return nil, err
	}

	log.Printf("Running pending migrations.\r\n")

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, err
	}

	game.LoadTerrain()
	game.LoadRaceTable()
	game.LoadJobTable()

	err = game.LoadSkills()
	if err != nil {
		return nil, err
	}

	game.world = make(map[uint]*Room)

	err = game.LoadZones()
	if err != nil {
		return nil, err
	}

	err = game.FixExits()
	if err != nil {
		return nil, err
	}

	err = game.InitScripting()
	if err != nil {
		return nil, err
	}

	err = game.LoadResets()
	if err != nil {
		return nil, err
	}

	return game, nil
}

/* Game loop */
func (game *Game) Run() {
	/* Handle violence logic */
	processCombatTicker := time.NewTicker(2 * time.Second)
	/* Handle frequent character update logic */
	processCharacterUpdateTicker := time.NewTicker(2 * time.Second)

	/* Buffered/paged output for clients */
	processOutputTicker := time.NewTicker(50 * time.Millisecond)

	processUpdateTicker := time.NewTicker(15 * time.Second)
	game.Update()

	/* Handle resets and trigger one immediately */
	processZoneUpdateTicker := time.NewTicker(1 * time.Minute)
	game.ZoneUpdate()

	game.doMazeTesting()

	for {
		select {
		case <-processUpdateTicker.C:
			game.Update()

		case <-processZoneUpdateTicker.C:
			game.ZoneUpdate()

		case <-processCharacterUpdateTicker.C:
			game.characterUpdate()

		case <-processCombatTicker.C:
			game.combatUpdate()

		case <-processOutputTicker.C:
			for client := range game.clients {
				if client.character != nil {
					if client.character.outputHead > 0 {
						client.displayPrompt()
					}

					client.character.flushOutput()
				}
			}

		case clientMessage := <-game.clientMessage:
			game.nanny(clientMessage.client, clientMessage.message)

		case client := <-game.register:
			game.clients[client] = true

			log.Printf("New connection from %s\r\n", client.conn.RemoteAddr().String())

			client.connectionState = ConnectionStateName

			client.send <- Config.greeting
			client.send <- []byte("By what name do you wish to be known? ")

		case client := <-game.unregister:
			delete(game.clients, client)

			if client.character != nil {
				log.Printf("Lost connection with %s@%s.\r\n", client.character.name, client.conn.RemoteAddr().String())

				client.character.client = nil
				break
			}

			log.Printf("Lost connection with %s.\r\n", client.conn.RemoteAddr().String())

		case quit := <-game.quitRequest:
			if quit.character != nil {
				quit.character.flushOutput()
			}

			quit.conn.Close()

		case <-game.shutdownRequest:
			os.Exit(0)
			return
		}
	}
}
