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
	"strings"
	"time"

	"github.com/dop251/goja"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

const (
	databaseDriverSQLite = "sqlite"
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

	eventHandlers   map[string]*LinkedList
	Scripts         map[uint]*Script `json:"scripts"`
	objectScripts   map[uint]*Script
	districtScripts map[int]*Script
	webhookScripts  map[int]*Script
	webhooks        map[string]*Webhook

	register                 chan *Client
	unregister               chan *Client
	quitRequest              chan *Client
	shutdownRequest          chan bool
	clientMessage            chan ClientTextMessage
	webhookMessage           chan string
	worldMapRequest          chan worldMapRequest
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
	game.worldMapRequest = make(chan worldMapRequest)
	game.clientMessage = make(chan ClientTextMessage)
	game.planeGenerationCompleted = make(chan int)

	game.Characters = NewLinkedList()
	game.Fights = NewLinkedList()
	game.Objects = NewLinkedList()
	game.ScriptTimers = NewLinkedList()
	game.Planes = NewLinkedList()

	/* Initialize services we'll inject elsewhere through the game instance. */
	game.db, err = openDatabase()
	if err != nil {
		return nil, err
	}

	/* Attempt new migrations at startup */
	err = runDatabaseMigrations(game.db)
	if err != nil {
		return nil, err
	}

	err = game.LoadTerrain()
	if err != nil {
		return nil, err
	}

	err = game.LoadRaceTable()
	if err != nil {
		return nil, err
	}

	err = game.LoadJobTable()
	if err != nil {
		return nil, err
	}

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

	/* Connect districts now that plane layers are initialized */
	err = game.LoadDistricts()
	if err != nil {
		return nil, err
	}

	err = game.LoadShops()
	if err != nil {
		return nil, err
	}

	err = game.LoadResets()
	if err != nil {
		return nil, err
	}

	/* Run district scripts */
	for districtId, script := range game.districtScripts {
		district := game.FindDistrictByID(districtId)
		if district == nil {
			log.Printf("Couldn't run district-script for nonexistent district id %d.\r\n", districtId)
			continue
		}

		_, err := script.tryEvaluate("onStart", game.vm.ToValue(district))
		if err != nil {
			log.Printf("Script evaluation of %d for district %d onStart failed: %v\r\n", script.Id, districtId, err)
		}
	}

	return game, nil
}

func openDatabase() (*sql.DB, error) {
	driverName, dsn, err := databaseConnectionInfo(Config.DatabaseConfiguration)
	if err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, err
	}

	configureDatabasePool(db)

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	err = configureDatabaseConnection(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func databaseConnectionInfo(config AppDatabaseConfiguration) (string, string, error) {
	driverName := strings.ToLower(strings.TrimSpace(config.Driver))
	if driverName == "" {
		driverName = databaseDriverSQLite
	}

	if driverName != databaseDriverSQLite {
		return "", "", fmt.Errorf("unsupported database driver %q; only sqlite is supported", config.Driver)
	}

	if config.DSN != "" {
		return driverName, config.DSN, nil
	}

	if config.Path == "" {
		return "", "", fmt.Errorf("sqlite database path must be configured")
	}

	return driverName, config.Path, nil
}

func configureDatabasePool(db *sql.DB) {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
}

func configureDatabaseConnection(db *sql.DB) error {
	_, err := db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return err
	}

	_, err = db.Exec("PRAGMA busy_timeout = 5000")
	return err
}

func runDatabaseMigrations(db *sql.DB) error {
	driver, databaseName, migrationSource, err := databaseMigrationDriver(db)
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(migrationSource, databaseName, driver)
	if err != nil {
		return err
	}

	log.Printf("Running pending migrations.\r\n")

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func databaseMigrationDriver(db *sql.DB) (database.Driver, string, string, error) {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	return driver, databaseDriverSQLite, "file://migrations", err
}

/* Game loop */
func (game *Game) Run() {
	/* Handle violence logic */
	processCombatTicker := time.NewTicker(2 * time.Second)

	/* Handle effect updates */
	processScriptTimersTicker := time.NewTicker(1 * time.Second)

	/* Handle frequent character update logic */
	processCharacterUpdateTicker := time.NewTicker(2 * time.Second)

	/* Handle object update logic */
	processObjectUpdateTicker := time.NewTicker(15 * time.Second)

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

		case <-processObjectUpdateTicker.C:
			game.objectUpdate()

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

		case request := <-game.worldMapRequest:
			response, err := game.buildWorldMapResponse()
			request.response <- worldMapResult{response: response, err: err}

		case client := <-game.register:
			game.clients[client] = true

			out := fmt.Sprintf("Network: new connection from %s\r\n", client.conn.RemoteAddr().String())
			log.Print(out)
			game.broadcast(out, WiznetBroadcastFilter)

			client.ConnectionState = ConnectionStateName

			client.Send(Config.greeting)
			client.Send([]byte("By what name do you wish to be known? "))

		case client := <-game.unregister:
			game.unregisterClient(client)

		case quit := <-game.quitRequest:
			if quit.Character != nil {
				quit.Character.flushOutput()
			}

			quit.Close()

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

func (game *Game) unregisterClient(client *Client) {
	delete(game.clients, client)

	var logOutput string

	if client.Character != nil {
		logOutput = fmt.Sprintf("Lost connection with %s@%s.\r\n", client.Character.Name, client.conn.RemoteAddr().String())

		if client.Character.Client == client {
			client.Character.Client = nil
		}

		log.Print(logOutput)
		game.broadcast(logOutput, WiznetBroadcastFilter)
		return
	}

	logOutput = fmt.Sprintf("Lost connection with %s.\r\n", client.conn.RemoteAddr().String())
	log.Print(logOutput)
	game.broadcast(logOutput, WiznetBroadcastFilter)
}
