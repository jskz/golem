package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"
)

/* Bust a prompt! */
func (client *Client) displayPrompt() {
	if client.character == nil {
		/* Something weird is going on: give a simple debug prompt */
		client.send <- []byte("\r\n> ")
		return
	}

	client.character.Write([]byte("\r\n> "))
}

func (game *Game) nanny(client *Client, message string) {
	var output bytes.Buffer

	/*
	 * The "nanny" handles line-based input from the client according to its connection state.
	 *
	 *
	 */
	switch client.connectionState {
	default:
		log.Printf("Client is trying to send a message from an invalid or unhandled connection state.\r\n")

	case ConnectionStatePlaying:
		client.character.Interpret(message)

	case ConnectionStateName:
		log.Printf("Guest attempting to login with name: %s\r\n", message)

		name := strings.Title(strings.ToLower(message))

		if !game.IsValidPCName(name) {
			output.WriteString("Invalid name, please try another.\r\n\r\nBy what name do you wish to be known? ")
			break
		}

		client.character = NewCharacter()
		client.character.client = client
		client.character.name = name
		client.connectionState = ConnectionStateConfirmName

		output.WriteString(fmt.Sprintf("No adventurer with that name exists.  Create %s? [y/N] ", client.character.name))

	case ConnectionStateConfirmName:
		if !strings.HasPrefix(strings.ToLower(message), "y") {
			client.connectionState = ConnectionStateName
			output.WriteString("\r\nBy what name do you wish to be known? ")
			break
		}

		client.connectionState = ConnectionStateNewPassword

		output.WriteString(fmt.Sprintf("Creating new character %s.\r\n", client.character.name))
		output.WriteString("Please choose a password: ")

	case ConnectionStateNewPassword:
		client.connectionState = ConnectionStateMessageOfTheDay

		output.WriteString("Bypassing password and character creation for very early development.\r\n\r\n")
		output.WriteString("[ Press return to join ]")

	case ConnectionStateMessageOfTheDay:
		client.connectionState = ConnectionStatePlaying
	}

	if client.connectionState != ConnectionStatePlaying && output.Len() > 0 {
		client.send <- output.Bytes()
	}
}
