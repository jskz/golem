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
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Game struct {
	db *sql.DB

	clients map[*Client]bool

	register      chan *Client
	unregister    chan *Client
	quitRequest   chan *Client
	clientMessage chan ClientTextMessage
}

func NewGame() (*Game, error) {
	var err error

	/* Create the game world instance and initialize variables & channels */
	game := &Game{}

	game.clients = make(map[*Client]bool)
	game.register = make(chan *Client)
	game.unregister = make(chan *Client)
	game.quitRequest = make(chan *Client)
	game.clientMessage = make(chan ClientTextMessage)

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

	return game, nil
}

/* Game loop */
func (game *Game) Run() {
	// processCombatTicker := time.NewTicker(2 * time.Second)
	processOutputTicker := time.NewTicker(50 * time.Millisecond)

	for {
		select {
		case <-processOutputTicker.C:
			for client := range game.clients {
				if client.character != nil {
					if client.character.pageCursor != 0 {
						client.displayPrompt()
					}

					client.character.flushOutput()
				}
			}

		case clientMessage := <-game.clientMessage:
			log.Printf("Received client message from %s: %s\r\n",
				clientMessage.client.conn.RemoteAddr().String(),
				clientMessage.message)

			game.nanny(clientMessage.client, clientMessage.message)

		case client := <-game.register:
			game.clients[client] = true

			log.Printf("New connection from %s.\r\n", client.conn.RemoteAddr().String())

			client.connectionState = ConnectionStateName
			client.send <- []byte("By what name do you wish to be known? ")

		case client := <-game.unregister:
			delete(game.clients, client)

			log.Printf("Lost connection with %s.\r\n", client.conn.RemoteAddr().String())

		case quit := <-game.quitRequest:
			if quit.character != nil {
				quit.character.flushOutput()
			}

			quit.conn.Close()
		}
	}
}
