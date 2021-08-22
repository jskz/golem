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
	"net"
)

func main() {
	/* Game instance will encapsulate both the world and player session management */
	game := NewGame()

	app, err := net.Listen("tcp", fmt.Sprintf(":%d", Config.Port))
	if err != nil {
		fmt.Println(err)
		return
	}

	/* Start the game loop */
	go game.Run()
	log.Printf("Golem is ready to rock and roll on port %d.\r\n", Config.Port)

	/* Spawn a new goroutine for each new client. */
	for {
		conn, err := app.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\r\n", err)
			continue
		}

		go game.handleConnection(conn)
	}
}
