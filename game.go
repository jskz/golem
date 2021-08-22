package main

import "log"

type Game struct {
	clients map[*Client]bool

	register      chan *Client
	unregister    chan *Client
	clientMessage chan ClientTextMessage
}

func NewGame() *Game {
	game := &Game{}

	game.clients = make(map[*Client]bool)
	game.register = make(chan *Client)
	game.unregister = make(chan *Client)
	game.clientMessage = make(chan ClientTextMessage)

	return game
}

/* Game loop */
func (game *Game) Run() {
	for {
		select {
		case clientMessage := <-game.clientMessage:
			log.Printf("Received client message from %s: %s\r\n",
				clientMessage.client.conn.RemoteAddr().String(),
				clientMessage.message)

			game.nanny(clientMessage.client, clientMessage.message)

		case client := <-game.register:
			game.clients[client] = true

			log.Printf("New connection from %s.\r\n", client.conn.RemoteAddr().String())

		case client := <-game.register:
			delete(game.clients, client)

			log.Printf("Lost connection with %s.\r\n", client.conn.RemoteAddr().String())
		}
	}
}
