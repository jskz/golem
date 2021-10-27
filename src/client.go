/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

/*
 * Telnet reference used: http://pcmicro.com/netfoss/telnet.html
 */

/* OOB telnet messaging */
const (
	TelnetECHO                 = 1
	TelnetSUPPRESSGOAHEAD      = 3
	TelnetTERMINALTYPE         = 24
	TelnetWINDOWSIZE           = 31
	TelnetTS                   = 32
	TelnetENVIRONMENTVARIABLES = 36
	TelnetNEWENVIRONMENT       = 39

	TelnetERASELINE = 248
	TelnetWILL      = 251
	TelnetWONT      = 252
	TelnetDO        = 253
	TelnetDONT      = 254
	TelnetIAC       = 255
)

type TelnetCommand struct {
	opCodes []byte
}

/* App-level connection state */
const (
	ConnectionStateNone            = 0
	ConnectionStateName            = 1
	ConnectionStateConfirmName     = 2
	ConnectionStatePassword        = 3
	ConnectionStateNewPassword     = 4
	ConnectionStateConfirmPassword = 5
	ConnectionStateChooseRace      = 6
	ConnectionStateConfirmRace     = 7
	ConnectionStateChooseClass     = 8
	ConnectionStateConfirmClass    = 9
	ConnectionStateMessageOfTheDay = 23
	ConnectionStatePlaying         = 24
)

/* Instance of a client connection */
type Client struct {
	id               string
	sessionStartedAt time.Time
	conn             net.Conn
	ansiEnabled      bool
	send             chan []byte
	close            chan bool
	character        *Character
	connectionState  uint
}

type ClientTextMessage struct {
	client  *Client
	message string
}

func (client *Client) readPump(game *Game) {
	defer func() {
		client.conn.Close()
	}()

	reader := bufio.NewReader(client.conn)

	for {
		firstByte, err := reader.Peek(1)
		if err != nil {
			log.Printf("Unable to read first byte: %v.\r\n", err)
			return
		}

		var commands []*TelnetCommand = []*TelnetCommand{}

		clientRequests := make([]byte, 0) /* IAC DO operation */
		clientWill := make([]byte, 0)     /* IAC WILL operation */

		if firstByte[0] == TelnetIAC {
			var nextByte []byte
			var length int = 3

			nextByte, err = reader.Peek(2)
			if nil != err {
				log.Printf("Unable to peek next byte after IAC: %v.\r\n", err)
				return
			}

			switch nextByte[0] {
			case TelnetDONT:
				length = 3
				requestOption, err := reader.Peek(3)
				if err != nil {
					log.Printf("Unable to peek next 2 bytes for IAC DONT: %v.\r\n", err)
					break
				}

				log.Printf("Client sent DONT %d.\r\n", requestOption[2])
			case TelnetWILL:
				length = 3

				willOption, err := reader.Peek(3)
				if err != nil {
					log.Printf("Unable to peek next 2 bytes for IAC WILL: %v.\r\n", err)
					break
				}

				clientWill = append(clientWill, willOption[2])
			case TelnetDO:
				length = 3

				requestOption, err := reader.Peek(3)
				if err != nil {
					log.Printf("Unable to peek next 2 bytes for IAC DO: %v.\r\n", err)
					break
				}

				clientRequests = append(clientWill, requestOption[2])

			/*
			 * To fix: I believe we are now only grabbing the first IAC command each time now - it still seems
			 * to passively work out, but should instead recursively peek ahead here for all commands at once.
			 */
			case TelnetIAC:
				break
			default:
				log.Printf("Unknown IAC code: %d.\r\n", nextByte[0])
			}

			reader.Discard(length)
		} else {
			line, err := reader.ReadString('\n')
			if err != nil {
				log.Printf("Failed to read string from reader: %v.\r\n", err)
				break
			}

			if len(line) > 512 {
				log.Printf("Client line input was too long, dropping connection.\r\n")
				return
			}

			trimmed := strings.TrimRight(line, "\r\n")

			clientMessage := ClientTextMessage{
				client:  client,
				message: trimmed,
			}

			game.clientMessage <- clientMessage
		}

		for _, command := range commands {
			if len(command.opCodes) < 3 {
				/* We can't make safe assumptions about this IAC command; skip. */
				continue
			}

			intent := command.opCodes[1]
			switch intent {
			case TelnetWILL:
				clientWill = append(clientWill, command.opCodes[2])
			case TelnetDO:
				clientRequests = append(clientRequests, command.opCodes[2])
			}
		}

		/*
		 * For every WILL/DO message the client has sent via telnet, we're going to respond
		 * with the appropriate "not supported, disable it" message until we incrementally
		 * add those protocol features.
		 */
		var telnetResponse bytes.Buffer

		for _, will := range clientWill {
			telnetResponse.Write([]byte{
				TelnetIAC,
				TelnetDONT,
				will,
			})
		}

		for _, do := range clientRequests {
			telnetResponse.Write([]byte{
				TelnetIAC,
				TelnetWONT,
				do,
			})
		}

		/* Only send this if necessary! */
		if telnetResponse.Len() > 0 {
			responseBytes := telnetResponse.Bytes()

			client.send <- responseBytes
		}
	}
}

func (client *Client) writePump(game *Game) {
	defer func() {
		close(client.send)

		game.unregister <- client
	}()

	for {
		select {
		case <-client.close:
			return

		case outgoing := <-client.send:
			_, err := client.conn.Write(outgoing)
			if err != nil {
				log.Printf("Error writing to socket: %v\r\n", err)
				return
			}
		}
	}
}

func (client *Client) closeConnection() {
	defer func() {
		recover()
	}()

	client.close <- true
}

func (game *Game) checkReconnect(client *Client, name string) bool {
	for iter := game.Characters.Head; iter != nil; iter = iter.Next {
		ch := iter.Value.(*Character)

		if ch.Flags&CHAR_IS_PLAYER != 0 && ch.Name == name {
			client.character = nil
			ch.Client = client

			client.character = ch
			client.connectionState = ConnectionStatePlaying

			ch.clearOutputBuffer()
			ch.Send("{MReconnecting to a session in progress.{x\r\n")

			if ch.Room != nil {
				for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
					character := iter.Value.(*Character)

					if character != ch {
						character.Send(fmt.Sprintf("\r\n{M%s has reconnected.{x\r\n", ch.GetShortDescriptionUpper(character)))
					}
				}
			}

			return true
		}
	}

	return false
}

func (game *Game) handleConnection(conn net.Conn) {
	defer func() {
		recover()
	}()

	client := &Client{sessionStartedAt: time.Now()}
	client.conn = conn
	client.send = make(chan []byte)
	client.close = make(chan bool)
	client.character = nil
	client.connectionState = ConnectionStateNone
	client.ansiEnabled = true

	/* Spawn two goroutines to handle client I/O */
	go client.readPump(game)
	go client.writePump(game)

	game.register <- client
}
