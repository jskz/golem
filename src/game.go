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
	"net/http"
	"os"
	"time"

	"github.com/dop251/goja"

	_ "github.com/go-sql-driver/mysql"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/mysql"
	_ "github.com/golang-migrate/migrate/source/file"
)

type Game struct {
	startedAt time.Time

	db *sql.DB
	vm *goja.Runtime

	Objects      *LinkedList `json:"objects"`
	Characters   *LinkedList `json:"characters"`
	Fights       *LinkedList `json:"fights"`
	Planes       *LinkedList `json:"planes"`
	Zones        *LinkedList `json:"zones"`
	ScriptTimers *LinkedList `json:"scriptTimers"`

	clients     map[*Client]bool
	skills      map[uint]*Skill
	world       map[uint]*Room
	shops       map[uint]*Shop
	mobileShops map[uint]*Shop

	eventHandlers  map[string]*LinkedList
	Scripts        map[uint]*Script `json:"scripts"`
	objectScripts  map[uint]*Script
	webhookScripts map[int]*Script
	webhooks       map[string]*Webhook

	register                 chan *Client
	unregister               chan *Client
	quitRequest              chan *Client
	shutdownRequest          chan bool
	clientMessage            chan ClientTextMessage
	webhookMessage           chan string
	planeGenerationCompleted chan int
}

func NewGame() (*Game, error) {
	var err error

	/* Start the profiler HTTP server if enabled */
	if Config.ProfilingConfiguration.Enabled {
		go func() {
			log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%d", Config.ProfilingConfiguration.Port), nil))
		}()
	}

	/* Create the game world instance and initialize variables & channels */
	game := &Game{startedAt: time.Now()}

	game.clients = make(map[*Client]bool)
	game.register = make(chan *Client)
	game.unregister = make(chan *Client)
	game.quitRequest = make(chan *Client)
	game.shutdownRequest = make(chan bool)
	game.webhookMessage = make(chan string)
	game.clientMessage = make(chan ClientTextMessage)
	game.planeGenerationCompleted = make(chan int)

	game.Characters = NewLinkedList()
	game.Fights = NewLinkedList()
	game.Objects = NewLinkedList()
	game.ScriptTimers = NewLinkedList()
	game.Planes = NewLinkedList()

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
		"file://migrations",
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

	err = game.LoadJobSkills()
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

	err = game.LoadPlanes()
	if err != nil {
		return nil, err
	}

	err = game.LoadWebhooks()
	if err != nil {
		return nil, err
	}

	err = game.InitScripting()
	if err != nil {
		return nil, err
	}

	/* Try to initialize each plane now that potential scripts have been attached */
	for iter := game.Planes.Head; iter != nil; iter = iter.Next {
		plane := iter.Value.(*Plane)

		log.Printf("Generating %s...\r\n", plane.Name)

		err = plane.generate()
		if err != nil {
			return nil, err
		}
	}

	err = game.LoadShops()
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

	/* Handle effect updates */
	processScriptTimersTicker := time.NewTicker(1 * time.Second)

	/* Handle frequent character update logic */
	processCharacterUpdateTicker := time.NewTicker(2 * time.Second)

	/* Buffered/paged output for clients */
	processOutputTicker := time.NewTicker(50 * time.Millisecond)

	processUpdateTicker := time.NewTicker(15 * time.Second)
	game.Update()

	/* Handle resets and trigger one immediately */
	processZoneUpdateTicker := time.NewTicker(1 * time.Minute)
	game.ZoneUpdate()

	for {
		select {
		case <-processUpdateTicker.C:
			game.Update()

		case <-processZoneUpdateTicker.C:
			game.ZoneUpdate()

		case <-processCharacterUpdateTicker.C:
			game.characterUpdate()

		case <-processScriptTimersTicker.C:
			game.scriptTimersUpdate()

		case <-processCombatTicker.C:
			game.combatUpdate()

		case <-processOutputTicker.C:
			for client := range game.clients {
				if client.Character != nil {
					if client.Character.outputHead > 0 {
						client.displayPrompt()
					}

					client.Character.flushOutput()
				}
			}

		case clientMessage := <-game.clientMessage:
			game.nanny(clientMessage.client, clientMessage.message)

		case webhookMessage := <-game.webhookMessage:
			webhook, ok := game.webhooks[webhookMessage]
			if !ok {
				log.Print("Received GET webhook request with a nonexistent key.\r\n")
				break
			}

			script, ok := game.webhookScripts[webhook.Id]
			if !ok {
				log.Print("Received GET webhook message for webhook without an attached script handler.\r\n")
				break
			}

			_, err := script.tryEvaluate("onGET", game.vm.ToValue(game))
			if err != nil {
				log.Printf("Script evaluation for webhook onGET request failed: %v\r\n", err)
			}

		case client := <-game.register:
			game.clients[client] = true

			out := fmt.Sprintf("Network: new connection from %s\r\n", client.conn.RemoteAddr().String())
			log.Print(out)
			game.broadcast(out, WiznetBroadcastFilter)

			client.ConnectionState = ConnectionStateName

			client.send <- Config.greeting
			client.send <- []byte("By what name do you wish to be known? ")

		case client := <-game.unregister:
			delete(game.clients, client)

			var logOutput string

			if client.Character != nil {
				logOutput = fmt.Sprintf("Lost connection with %s@%s.\r\n", client.Character.Name, client.conn.RemoteAddr().String())

				client.Character.Client = nil
				log.Print(logOutput)
				game.broadcast(logOutput, WiznetBroadcastFilter)
				break
			}

			logOutput = fmt.Sprintf("Lost connection with %s.\r\n", client.conn.RemoteAddr().String())
			log.Print(logOutput)
			game.broadcast(logOutput, WiznetBroadcastFilter)

		case quit := <-game.quitRequest:
			if quit.Character != nil {
				quit.Character.flushOutput()
			}

			quit.conn.Close()

		case planeId := <-game.planeGenerationCompleted:
			plane := game.FindPlaneByID(planeId)
			if plane != nil {
				if plane.Scripts != nil {
					plane.Scripts.tryEvaluate("onGenerationComplete", plane.Game.vm.ToValue(game), plane.Game.vm.ToValue(plane))
				}
			}

		case <-game.shutdownRequest:
			os.Exit(0)
			return
		}
	}
}
