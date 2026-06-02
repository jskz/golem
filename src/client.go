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
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
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

	TelnetENDSUBNEGOTIATION = 240
	TelnetNOP               = 241
	TelnetDATAMARK          = 242
	TelnetBREAK             = 243
	TelnetINTERRUPTPROCESS  = 244
	TelnetABORTOUTPUT       = 245
	TelnetAREYOUTHERE       = 246
	TelnetERASECHARACTER    = 247
	TelnetERASELINE         = 248
	TelnetGOAHEAD           = 249
	TelnetSUBNEGOTIATION    = 250
	TelnetWILL              = 251
	TelnetWONT              = 252
	TelnetDO                = 253
	TelnetDONT              = 254
	TelnetIAC               = 255
)

const (
	clientMaxLineLength  = 512
	clientSendBufferSize = 64
	clientWriteTimeout   = 10 * time.Second
)

var errClientLineTooLong = errors.New("client line input was too long")

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
	ConnectionStateRollingStats    = 10
	ConnectionStateMessageOfTheDay = 23
	ConnectionStatePlaying         = 24
	ConnectionStateMax             = 25
)

/* Instance of a client connection */
type Client struct {
	sessionStartedAt  time.Time
	conn              net.Conn
	ansiEnabled       bool
	send              chan []byte
	close             chan struct{}
	closeOnce         sync.Once
	unregisterOnce    sync.Once
	remainingRolls    int
	delayMutex        sync.Mutex
	delayUntil        time.Time
	Character         *Character     `json:"character"`
	ConnectionState   uint           `json:"connectionState"`
	ConnectionHandler *goja.Callable `json:"connectionHandler"`
}

type ClientTextMessage struct {
	client  *Client
	message string
}

func (client *Client) readPump(game *Game) {
	defer func() {
		client.Close()
		client.unregister(game)
	}()

	reader := newClientInputReader(client.conn)

	for {
		firstByte, err := reader.Peek(1)
		if err != nil {
			if err == io.EOF {
				return
			}

			log.Printf("Unable to read first byte: %v.\r\n", err)
			return
		}

		clientRequests := make([]byte, 0) /* IAC DO operation */
		clientWill := make([]byte, 0)     /* IAC WILL operation */

		if firstByte[0] == TelnetIAC {
			clientWill, clientRequests, err = handleTelnetCommand(reader)
			if err != nil {
				log.Printf("Unable to handle IAC command: %v.\r\n", err)
				return
			}
		} else {
			trimmed, err := readClientLine(reader)
			if err == errClientLineTooLong {
				log.Printf("Client line input was too long, dropping connection.\r\n")
				return
			}

			if err != nil {
				log.Printf("Failed to read string from reader: %v.\r\n", err)
				break
			}

			clientMessage := ClientTextMessage{
				client:  client,
				message: trimmed,
			}

			client.delayMutex.Lock()
			delay := client.delayUntil
			client.delayMutex.Unlock()

			if time.Now().Before(delay) {
				<-time.After(time.Duration(time.Until(delay)))
			}

			game.clientMessage <- clientMessage
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

			if client.Send(responseBytes) {
				break
			}
		}
	}
}

func handleTelnetCommand(reader *bufio.Reader) ([]byte, []byte, error) {
	clientWill := make([]byte, 0)
	clientRequests := make([]byte, 0)

	nextByte, err := reader.Peek(2)
	if err != nil {
		return clientWill, clientRequests, fmt.Errorf("peek next byte after IAC: %w", err)
	}

	switch nextByte[1] {
	case TelnetDONT:
		option, err := readTelnetNegotiation(reader)
		if err != nil {
			return clientWill, clientRequests, fmt.Errorf("read IAC DONT option: %w", err)
		}

		log.Printf("Client sent DONT %d.\r\n", option)
	case TelnetWONT:
		option, err := readTelnetNegotiation(reader)
		if err != nil {
			return clientWill, clientRequests, fmt.Errorf("read IAC WONT option: %w", err)
		}

		log.Printf("Client sent WONT %d.\r\n", option)
	case TelnetWILL:
		option, err := readTelnetNegotiation(reader)
		if err != nil {
			return clientWill, clientRequests, fmt.Errorf("read IAC WILL option: %w", err)
		}

		clientWill = append(clientWill, option)
	case TelnetDO:
		option, err := readTelnetNegotiation(reader)
		if err != nil {
			return clientWill, clientRequests, fmt.Errorf("read IAC DO option: %w", err)
		}

		clientRequests = append(clientRequests, option)
	case TelnetSUBNEGOTIATION:
		err = discardTelnetSubnegotiation(reader)
	case TelnetENDSUBNEGOTIATION,
		TelnetNOP,
		TelnetDATAMARK,
		TelnetBREAK,
		TelnetINTERRUPTPROCESS,
		TelnetABORTOUTPUT,
		TelnetAREYOUTHERE,
		TelnetERASECHARACTER,
		TelnetERASELINE,
		TelnetGOAHEAD,
		TelnetIAC:
		_, err = reader.Discard(2)
	default:
		log.Printf("Unknown IAC code: %d.\r\n", nextByte[1])
		_, err = reader.Discard(2)
	}

	return clientWill, clientRequests, err
}

func readTelnetNegotiation(reader *bufio.Reader) (byte, error) {
	opCodes, err := reader.Peek(3)
	if err != nil {
		return 0, err
	}

	option := opCodes[2]
	_, err = reader.Discard(3)

	return option, err
}

func discardTelnetSubnegotiation(reader *bufio.Reader) error {
	if _, err := reader.Discard(2); err != nil {
		return err
	}

	for {
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}

		if b != TelnetIAC {
			continue
		}

		command, err := reader.ReadByte()
		if err != nil {
			return err
		}

		if command == TelnetENDSUBNEGOTIATION {
			return nil
		}
	}
}

func newClientInputReader(input io.Reader) *bufio.Reader {
	return bufio.NewReaderSize(input, clientMaxLineLength+1)
}

func readClientLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		return "", errClientLineTooLong
	}

	if err != nil {
		return "", err
	}

	if len(line) > clientMaxLineLength {
		return "", errClientLineTooLong
	}

	return strings.TrimRight(string(line), "\r\n"), nil
}

func (client *Client) writePump(game *Game) {
	defer func() {
		client.Close()

		client.unregister(game)
	}()

	for {
		select {
		case <-client.close:
			return

		case outgoing, ok := <-client.send:
			if !ok {
				return
			}

			if err := client.writeToConn(outgoing); err != nil {
				log.Printf("Error writing to socket: %v\r\n", err)
				return
			}
		}
	}
}

func (client *Client) writeToConn(outgoing []byte) error {
	for len(outgoing) > 0 {
		err := client.conn.SetWriteDeadline(time.Now().Add(clientWriteTimeout))
		if err != nil {
			return err
		}

		n, err := client.conn.Write(outgoing)
		if err != nil {
			return err
		}

		if n == 0 {
			return io.ErrShortWrite
		}

		outgoing = outgoing[n:]
	}

	return nil
}

func (client *Client) Close() {
	client.closeOnce.Do(func() {
		close(client.close)
		client.conn.Close()
	})
}

func (client *Client) unregister(game *Game) {
	client.unregisterOnce.Do(func() {
		game.unregister <- client
	})
}

func (client *Client) Send(data []byte) (closed bool) {
	select {
	case <-client.close:
		return true
	default:
	}

	outgoing := make([]byte, len(data))
	copy(outgoing, data)

	select {
	case <-client.close:
		return true
	case client.send <- outgoing:
		return false
	default:
		log.Printf("Client send queue full, dropping connection.\r\n")
		client.Close()
		return true
	}
}

func (game *Game) checkReconnect(client *Client, name string) bool {
	for iter := game.Characters.Head; iter != nil; iter = iter.Next {
		ch := iter.Value.(*Character)

		if ch.Flags&CHAR_IS_PLAYER != 0 && ch.Name == name {
			client.Character = nil
			ch.Client = client

			client.Character = ch
			client.ConnectionState = ConnectionStatePlaying

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

func (client *Client) Delay(ms int) {
	client.delayMutex.Lock()
	client.delayUntil = time.Now().Add(time.Duration(ms) * time.Millisecond)
	client.delayMutex.Unlock()
}

func (game *Game) handleConnection(conn net.Conn) {
	defer func() {
		recover()
	}()

	client := &Client{sessionStartedAt: time.Now()}
	client.conn = conn
	client.send = make(chan []byte, clientSendBufferSize)
	client.close = make(chan struct{})
	client.Character = nil
	client.remainingRolls = 10
	client.ConnectionState = ConnectionStateNone
	client.delayUntil = time.Now()
	client.delayMutex = sync.Mutex{}
	client.ansiEnabled = true

	/* Spawn goroutines to handle client I/O */
	go client.writePump(game)

	game.register <- client

	go client.readPump(game)
}
