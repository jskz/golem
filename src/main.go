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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"syscall"
	"time"

	_ "net/http/pprof"

	"golang.org/x/sys/unix"
)

const (
	CopyoverDataPath = "copyover.json"
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

	/* Attempt read of copyover sessions */
	copyoverBytes, err := ioutil.ReadFile(CopyoverDataPath)
	if err != nil {
		/* Don't try to restore these copyover sessions. */
		log.Println(err)
	} else {
		sessions := &CopyoverData{}

		err = json.Unmarshal(copyoverBytes, &sessions)
		if err != nil {
			log.Printf("Failed to unmarshal previous copyover data: %v\r\n", err)
			return
		}

		/* Initialize a client, attach the FD saved in the fd, pre-heat the object instances for that user, set playing */
		for _, session := range sessions.Sessions {
			client := &Client{}
			client.conn, err = net.FileConn(os.NewFile(uintptr(session.Fd), session.Name))

			character, _, err := game.FindPlayerByName(session.Name)
			if err != nil {
				log.Println(err)
				panic(err)
			}

			room, err := game.LoadRoomIndex(uint(session.Room))
			if err != nil {
				log.Println(err)
				panic(err)
			}

			room.AddCharacter(character)
			game.Characters.Insert(character)

			client.Character = character
			client.Character.Client = client
			client.ConnectionState = ConnectionStatePlaying

			log.Printf("Hot loaded %s back to room %d after a copyover.\r\n", character.Name, room.Id)
		}

		err = os.Remove(CopyoverDataPath)
		if err != nil {
			log.Printf("Failed to unlink copyover data after loading it: %v\r\n", err)
			return
		}
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
