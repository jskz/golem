/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	_ "net/http/pprof"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	/* Game instance will encapsulate both the world and player session management */
	game, err := NewGame()
	if err != nil {
		log.Printf("Unable to initialize new game session: %v.\r\n", err)
		os.Exit(1)
	}

	app, err := net.Listen("tcp", fmt.Sprintf(":%d", Config.Port))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	/* Spawn the webhook-handling goroutine */
	go game.handleWebhooks()

	/* Start the game loop */
	go game.Run()

	log.Printf("Golem is ready to rock and roll on port %d.\r\n", Config.Port)

	/* Spawn a new goroutine for each new client. */
	for {
		conn, err := app.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v\r\n", err)
			continue
		}

		go game.handleConnection(conn)
	}
}
