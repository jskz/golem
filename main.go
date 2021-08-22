package main

import (
	"fmt"
	"log"
	"net"
)

func main() {
	game := NewGame()

	app, err := net.Listen("tcp", ":4000")
	if err != nil {
		fmt.Println(err)
		return
	}

	go game.Run()

	log.Printf("Golem is ready to rock and roll on port 4000.\r\n")

	for {
		conn, err := app.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection: %v\r\n", err)
			continue
		}

		go game.handleConnection(conn)
	}
}
