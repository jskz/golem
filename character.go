package main

import "log"

/*
 * This character structure is shared by both player-characters (human beings
 * connected through a session instance available via the client pointer.)
 */
type Character struct {
	client     *Client
	pages      [][]byte
	pageSize   int
	pageCursor int

	name             string
	level            int
	shortDescription string
	longDescription  string
}

func (ch *Character) flushOutput() {
	for _, page := range ch.pages {
		ch.client.send <- page
	}

	ch.pages = make([][]byte, 1)
	ch.pages[0] = make([]byte, ch.pageSize)
	ch.pageCursor = 0
}

func (ch *Character) Write(data []byte) (n int, err error) {
	if ch.client == nil {
		/* If there is no client, succeed silently. */
		return len(data), nil
	}

	copy(ch.pages[ch.pageCursor/ch.pageSize][ch.pageCursor:ch.pageCursor+len(data)], data[:])
	ch.pageCursor = ch.pageCursor + len(data)

	return len(data), nil
}

/*
 * TODO: implement validation logic restricting silly/invalid/breaking names.
 */
func (game *Game) IsValidPCName(name string) bool {
	/* Length bounds */
	if len(name) < 3 || len(name) > 14 {
		return false
	}

	/* If any character is non-alpha, invalidate. */
	for c := range name {
		if ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') {
			return false
		}
	}

	/* TODO: entity checking; does a persistent player share this valid name? */
	return true
}

func (ch *Character) send(text string) {
	/*
	 * Mock implementation:
	 *
	 * We'll want to implement paging and allow the telnet protocol to negotiate
	 * the window size, or else make this configurable within the game settings.
	 */
	n, err := ch.Write([]byte(text))
	if err != nil {
		log.Printf("Failed to write to character: %v.\r\n", err)
		return
	}

	log.Printf("Successfully wrote %d to character buffer.\r\n", n)
}

func NewCharacter() *Character {
	character := &Character{}
	character.pageSize = 1024
	character.pages = make([][]byte, 1)
	character.pages[0] = make([]byte, character.pageSize)
	character.pageCursor = 0

	character.name = "formless protoplasm"
	character.client = nil
	character.level = 0

	return character
}
