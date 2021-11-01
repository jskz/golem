/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"syscall"
	"time"

	_ "net/http/pprof"

	"golang.org/x/sys/unix"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	/* Game instance will encapsulate both the world and player session management */
	game, err := NewGame()
	if err != nil {
		log.Printf("Unable to initialize new game session: %v.\r\n", err)
		os.Exit(1)
	}

	listenConfig := net.ListenConfig{
		Control: func(network, address string, conn syscall.RawConn) error {
			var err error = nil

			conn.Control(func(fd uintptr) {
				err = syscall.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			})

			return err
		},
	}

	app, err := listenConfig.Listen(context.Background(), "tcp", fmt.Sprintf(":%d", Config.Port))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

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
