/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"errors"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"
)

func main() {
	os.Exit(run())
}

func run() int {
	copyoverState, err := copyoverStateFromEnvironment()
	if err != nil {
		log.Printf("Unable to read copyover state: %v.\r\n", err)
		return 1
	}

	/* Game instance will encapsulate both the world and player session management */
	game, err := NewGame()
	if err != nil {
		log.Printf("Unable to initialize new game session: %v.\r\n", err)
		return 1
	}

	app, err := openListener(copyoverState)
	if err != nil {
		log.Println(err)
		return 1
	}
	game.listener = app
	defer app.Close()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(shutdown)

	go func() {
		<-shutdown
		log.Printf("Shutdown signal received.\r\n")
		app.Close()
	}()

	/* Spawn the webhook-handling goroutine */
	go game.handleWebhooks()

	if copyoverState != nil {
		err = game.recoverCopyover(copyoverState)
		if err != nil {
			log.Printf("Unable to recover copyover clients: %v.\r\n", err)
			return 1
		}
	}

	/* Start the game loop */
	go game.Run()

	log.Printf("Golem is ready to rock and roll on port %d.\r\n", Config.Port)

	/* Spawn a new goroutine for each new client. */
	for {
		conn, err := app.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return 0
			}

			log.Printf("Failed to accept connection: %v\r\n", err)
			continue
		}

		go game.handleConnection(conn)
	}
}
