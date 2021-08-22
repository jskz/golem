package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	/* Game instance will encapsulate both the world and player session management */
	game := NewGame()

	/* TODO: make port configurable :) */
	app, err := net.Listen("tcp", ":4000")
	if err != nil {
		fmt.Println(err)
		return
	}

	/* Start the game loop */
	go game.Run()
	log.Printf("Golem is ready to rock and roll on port 4000.\r\n")

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
