/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const JoinedGameFlavourText = "{WYou have entered the world of Golem.{x"

/* Bust a prompt! */
func (client *Client) displayPrompt() {
	if client.character == nil {
		/* Something weird is going on: give a simple debug prompt */
		client.send <- []byte("\r\n> ")
		return
	}

	var prompt bytes.Buffer

	/*
	 * TODO: if the character's paging cursor is in a page less than the top page, instead display pager info.
	 */
	healthPercentage := client.character.health * 100 / client.character.maxHealth
	manaPercentage := client.character.mana * 100 / client.character.maxMana
	staminaPercentage := client.character.stamina * 100 / client.character.maxStamina

	currentHealthColour := SeverityColourFromPercentage(healthPercentage)
	currentManaColour := SeverityColourFromPercentage(manaPercentage)
	currentStaminaColour := SeverityColourFromPercentage(staminaPercentage)

	prompt.WriteString(
		client.TranslateColourCodes(fmt.Sprintf("\r\n{w[%s%d{w/{G%d{ghp %s%d{w/{G%d{gm %s%d{w/{G%d{gst{w]{x ",
			currentHealthColour,
			client.character.health,
			client.character.maxHealth,
			currentManaColour,
			client.character.mana,
			client.character.maxMana,
			currentStaminaColour,
			client.character.stamina,
			client.character.maxStamina)))
	client.character.Write(prompt.Bytes())
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

	case ConnectionStatePassword:
		if !game.AttemptLogin(client.character.name, message) {
			client.connectionState = ConnectionStateName
			client.character = nil

			output.WriteString("Wrong password.\r\n\r\nBy what name do you wish to be known? ")
			break
		}

		for other := range game.clients {
			if other != client && other.character != nil && other.character.name == client.character.name {
				delete(game.clients, other)

				other.conn.Close()
			}
		}

		if game.checkReconnect(client, client.character.name) {
			break
		}

		client.connectionState = ConnectionStateMessageOfTheDay
		output.WriteString(string(Config.motd))
		output.WriteString("[ Press return to continue ]")

	case ConnectionStateName:
		log.Printf("Guest attempting to login with name: %s\r\n", message)

		name := strings.Title(strings.ToLower(message))

		if !game.IsValidPCName(name) {
			output.WriteString("Invalid name, please try another.\r\n\r\nBy what name do you wish to be known? ")
			break
		}

		character, room, err := game.FindPlayerByName(name)
		if err != nil {
			panic(err)
		}

		if character != nil {
			client.character = character
			client.character.flags |= CHAR_IS_PLAYER
			client.character.client = client
			output.WriteString("Password: ")
			client.character.room = room
			client.connectionState = ConnectionStatePassword
			break
		}

		client.character = NewCharacter()
		client.character.game = game
		client.character.client = client
		client.character.name = name
		client.character.level = 1
		client.character.flags |= CHAR_IS_PLAYER
		client.connectionState = ConnectionStateConfirmName

		client.character.practices = 100
		client.character.strength = 10
		client.character.dexterity = 10
		client.character.intelligence = 10
		client.character.wisdom = 10
		client.character.constitution = 10
		client.character.charisma = 10
		client.character.luck = 10

		client.character.health = 20
		client.character.maxHealth = 20

		client.character.mana = 100
		client.character.maxMana = 100

		client.character.stamina = 100
		client.character.maxStamina = 100

		output.WriteString(fmt.Sprintf("No adventurer with that name exists.  Create %s? [y/N] ", client.character.name))

	case ConnectionStateConfirmName:
		if !strings.HasPrefix(strings.ToLower(message), "y") {
			client.connectionState = ConnectionStateName
			client.character.name = UnauthenticatedUsername
			output.WriteString("\r\nBy what name do you wish to be known? ")
			break
		}

		client.connectionState = ConnectionStateNewPassword

		output.WriteString(fmt.Sprintf("Creating new character %s.\r\n", client.character.name))
		output.WriteString("Please choose a password: ")

	case ConnectionStateNewPassword:
		client.connectionState = ConnectionStateConfirmPassword
		ciphertext, err := bcrypt.GenerateFromPassword([]byte(message), 8)
		if err != nil {
			log.Println("Failed to bcrypt user password: ", err)
			return
		}

		client.character.temporaryHash = string(ciphertext)
		output.WriteString("Please confirm your password: ")

	case ConnectionStateConfirmPassword:
		if bcrypt.CompareHashAndPassword([]byte(client.character.temporaryHash), []byte(message)) != nil {
			client.connectionState = ConnectionStateNewPassword
			output.WriteString("Passwords didn't match.\r\nPlease choose a password: ")
			break
		}

		client.connectionState = ConnectionStateChooseRace
		output.WriteString("Please choose a race from the following options:\r\n")

		/* Counter value for periodically line-breaking */
		index := 0

		for iter := Races.head; iter != nil; iter = iter.next {
			race := iter.value.(*Race)

			if !race.Playable {
				continue
			}

			output.WriteString(fmt.Sprintf("%-12s ", race.Name))

			index++

			if index%7 == 0 {
				output.WriteString("\r\n")
			}
		}

		output.WriteString("\r\nChoice: ")

	case ConnectionStateChooseRace:
		race := FindRaceByName(message)
		if race == nil || !race.Playable {
			output.WriteString("\r\nInvalid choice for race, please choose another: ")
			break
		}

		client.character.race = race
		client.connectionState = ConnectionStateConfirmRace
		output.WriteString(fmt.Sprintf("\r\nAre you sure you want to be a %s? [y/N] ", race.Name))

	case ConnectionStateConfirmRace:
		if !strings.HasPrefix(strings.ToLower(message), "y") {
			client.connectionState = ConnectionStateChooseRace
			output.WriteString("Please choose a race from the following options:\r\n")

			/* Counter value for periodically line-breaking */
			index := 0

			for iter := Races.head; iter != nil; iter = iter.next {
				race := iter.value.(*Race)

				if !race.Playable {
					continue
				}

				output.WriteString(fmt.Sprintf("%-12s ", race.Name))

				index++

				if index%7 == 0 {
					output.WriteString("\r\n")
				}
			}

			output.WriteString("\r\nChoice: ")
			break
		}

		client.connectionState = ConnectionStateChooseClass
		output.WriteString("\r\nPlease choose a job from the following options:\r\n")

		/* Counter value for periodically line-breaking */
		index := 0

		for iter := Jobs.head; iter != nil; iter = iter.next {
			job := iter.value.(*Job)

			if !job.Playable {
				continue
			}

			output.WriteString(fmt.Sprintf("%-12s ", job.Name))

			index++

			if index%7 == 0 {
				output.WriteString("\r\n")
			}
		}

		output.WriteString("\r\nChoice: ")

	case ConnectionStateChooseClass:
		job := FindJobByName(message)
		if job == nil || !job.Playable {
			output.WriteString("\r\nInvalid choice for job, please choose another: ")
			break
		}

		client.character.job = job
		client.connectionState = ConnectionStateConfirmClass
		output.WriteString(fmt.Sprintf("\r\nAre you sure you want to be a %s? [y/N] ", job.Name))

	case ConnectionStateConfirmClass:
		if !strings.HasPrefix(strings.ToLower(message), "y") {
			client.connectionState = ConnectionStateChooseClass
			output.WriteString("Please choose a job from the following options:\r\n")

			/* Counter value for periodically line-breaking */
			index := 0

			for iter := Jobs.head; iter != nil; iter = iter.next {
				job := iter.value.(*Job)

				if !job.Playable {
					continue
				}

				output.WriteString(fmt.Sprintf("%-12s ", job.Name))
				index++

				if index%7 == 0 {
					output.WriteString("\r\n")
				}
			}

			output.WriteString("\r\nChoice: ")
			break
		}

		if !client.character.Finalize() {
			log.Printf("Unable to create new character %v, dropping connection.\r\n", client.character)
			client.conn.Close()
			break
		}

		client.connectionState = ConnectionStateMessageOfTheDay
		output.WriteString(string(Config.motd))
		output.WriteString("[ Press return to continue ]")

	case ConnectionStateMessageOfTheDay:
		client.connectionState = ConnectionStatePlaying

		game.characters.Insert(client.character)

		if client.character.room != nil {
			client.character.room.addCharacter(client.character)

			for iter := client.character.room.characters.head; iter != nil; iter = iter.next {
				character := iter.value.(*Character)

				if character != client.character {
					character.Send(fmt.Sprintf("{W%s has entered the game.{x\r\n", client.character.name))
				}
			}
		}

		client.character.Send(fmt.Sprintf("%s\r\n", JoinedGameFlavourText))
		do_look(client.character, "")
	}

	if client.connectionState != ConnectionStatePlaying && output.Len() > 0 {
		client.send <- output.Bytes()
	}
}
