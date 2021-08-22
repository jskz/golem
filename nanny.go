package main

import "log"

func (game *Game) nanny(client *Client, message string) {
	switch client.connectionState {
	default:
		log.Printf("Client is trying to send a message from an invalid connection state.\r\n")

	case ConnectionStatePlaying:
		log.Printf("Received gameplay input: %s.\r\n", message)
	}
}
