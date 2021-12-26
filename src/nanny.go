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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/getsentry/sentry-go"
	"golang.org/x/crypto/bcrypt"
)

const JoinedGameFlavourText = "{WYou have entered the world of Golem.{x"
const DefaultMaxLines = 50

/* Bust a prompt! */
func (client *Client) displayPrompt() {
	if client.Character == nil {
		/* Something weird is going on: give a simple debug prompt */
		client.Send([]byte("\r\n> "))
		return
	}

	if client.ConnectionState == ConnectionStateNone {
		return
	}

	var prompt bytes.Buffer
	if client.Character.outputCursor >= DefaultMaxLines && client.Character.inputCursor >= DefaultMaxLines {
		return
	}

	healthPercentage := client.Character.Health * 100 / client.Character.MaxHealth
	manaPercentage := client.Character.Mana * 100 / client.Character.MaxMana
	staminaPercentage := client.Character.Stamina * 100 / client.Character.MaxStamina

	currentHealthColour := SeverityColourFromPercentage(healthPercentage)
	currentManaColour := SeverityColourFromPercentage(manaPercentage)
	currentStaminaColour := SeverityColourFromPercentage(staminaPercentage)

	prompt.WriteString("\r\n")

	if client.Character.isAdmin() && client.Character.Room != nil && client.Character.Room.Flags&ROOM_PLANAR != 0 {
		prompt.WriteString(client.TranslateColourCodes(fmt.Sprintf("{C(%d, %d)", client.Character.Room.X, client.Character.Room.Y)))
	}

	if client.Character.Room != nil && client.Character.Room.Flags&ROOM_SAFE != 0 {
		prompt.WriteString(client.TranslateColourCodes("{W[SAFE]"))
	}

	prompt.WriteString(
		client.TranslateColourCodes(fmt.Sprintf("{w[%s%d{w/{G%d{ghp %s%d{w/{G%d{gm %s%d{w/{G%d{gst{w]{x ",
			currentHealthColour,
			client.Character.Health,
			client.Character.MaxHealth,
			currentManaColour,
			client.Character.Mana,
			client.Character.MaxMana,
			currentStaminaColour,
			client.Character.Stamina,
			client.Character.MaxStamina)))
	client.Character.Write(prompt.Bytes())
}

func (game *Game) nanny(client *Client, message string) {
	var output bytes.Buffer

	if Config.SentryConfiguration.Enabled {
		defer sentry.Recover()

		sentry.ConfigureScope(func(scope *sentry.Scope) {
			ctx := make(map[string]interface{})

			ctx["remote_address"] = client.conn.RemoteAddr().String()

			if client.Character != nil {
				ctx["name"] = client.Character.Name
				ctx["id"] = client.Character.Id

				if client.Character.Room != nil {
					ctx["room_id"] = client.Character.Room.Id
					ctx["room_name"] = client.Character.Room.Name
				}
			}

			ctx["input"] = message

			scope.SetContext("client", ctx)
		})
	}

	/*
	 * The "nanny" handles line-based input from the client according to its connection state.
	 *
	 *
	 */
	switch client.ConnectionState {
	default:
		log.Printf("Client is trying to send a message from an invalid or unhandled connection state.\r\n")

	case ConnectionStatePlaying:
		client.Character.Interpret(message)

	case ConnectionStatePassword:
		if !game.AttemptLogin(client.Character.Name, message) {
			client.ConnectionState = ConnectionStateName
			client.Character = nil

			output.WriteString("Wrong password.\r\n\r\nBy what name do you wish to be known? ")
			break
		}

		for other := range game.clients {
			if other != client && other.Character != nil && other.Character.Name == client.Character.Name {
				delete(game.clients, other)

				other.conn.Close()
			}
		}

		if game.checkReconnect(client, client.Character.Name) {
			break
		}

		client.Character.Client = client
		client.ConnectionState = ConnectionStateMessageOfTheDay
		output.WriteString(string(Config.motd))
		output.WriteString("[ Press return to continue ]")

	case ConnectionStateName:
		name := strings.Title(strings.ToLower(message))
		if !game.IsValidPCName(name) {
			output.WriteString("Invalid name, please try another.\r\n\r\nBy what name do you wish to be known? ")
			break
		}

		out := fmt.Sprintf("Guest attempting to login with name: %s\r\n", name)
		log.Print(out)
		game.broadcast(out, WiznetBroadcastFilter)

		character, room, err := game.FindPlayerByName(name)
		if err != nil {
			panic(err)
		}

		if character != nil {
			client.Character = character
			client.Character.Flags |= CHAR_IS_PLAYER
			output.WriteString("Password: ")
			client.Character.Room = room
			client.ConnectionState = ConnectionStatePassword
			break
		}

		client.Character = NewCharacter()
		client.Character.Game = game
		client.Character.Client = client
		client.Character.Name = name
		client.Character.Level = 1
		client.Character.Flags |= CHAR_IS_PLAYER
		client.ConnectionState = ConnectionStateConfirmName

		client.Character.Practices = 100

		for index, _ := range client.Character.Stats {
			client.Character.Stats[index] = 10
		}

		client.Character.Health = 20
		client.Character.MaxHealth = 20

		client.Character.Mana = 100
		client.Character.MaxMana = 100

		client.Character.Stamina = 100
		client.Character.MaxStamina = 100

		output.WriteString(fmt.Sprintf("No adventurer with that name exists.  Create %s? [y/N] ", client.Character.Name))

	case ConnectionStateConfirmName:
		if !strings.HasPrefix(strings.ToLower(message), "y") {
			client.ConnectionState = ConnectionStateName
			client.Character.Name = UnauthenticatedUsername
			output.WriteString("\r\nBy what name do you wish to be known? ")
			break
		}

		client.ConnectionState = ConnectionStateNewPassword

		output.WriteString(fmt.Sprintf("Creating new character %s.\r\n", client.Character.Name))
		output.WriteString("Please choose a password: ")

	case ConnectionStateNewPassword:
		client.ConnectionState = ConnectionStateConfirmPassword

		sha256Sum := sha256.Sum256([]byte(message + Config.HashSalt))
		saltedHash := hex.EncodeToString(sha256Sum[:])
		ciphertext, err := bcrypt.GenerateFromPassword([]byte(saltedHash), 10)
		if err != nil {
			log.Println("Failed to bcrypt user password: ", err)
			return
		}

		client.Character.temporaryHash = string(ciphertext)
		output.WriteString("Please confirm your password: ")

	case ConnectionStateConfirmPassword:
		sha256Sum := sha256.Sum256([]byte(message + Config.HashSalt))
		saltedHash := hex.EncodeToString(sha256Sum[:])

		if bcrypt.CompareHashAndPassword([]byte(client.Character.temporaryHash), []byte(saltedHash)) != nil {
			client.ConnectionState = ConnectionStateNewPassword
			output.WriteString("Passwords didn't match.\r\nPlease choose a password: ")
			break
		}

		client.ConnectionState = ConnectionStateChooseRace
		output.WriteString("Please choose a race from the following options:\r\n")

		/* Counter value for periodically line-breaking */
		index := 0

		for iter := Races.Head; iter != nil; iter = iter.Next {
			race := iter.Value.(*Race)

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

		client.Character.Race = race
		client.ConnectionState = ConnectionStateConfirmRace
		output.WriteString(fmt.Sprintf("\r\nAre you sure you want to be a %s? [y/N] ", race.Name))

	case ConnectionStateConfirmRace:
		if !strings.HasPrefix(strings.ToLower(message), "y") {
			client.ConnectionState = ConnectionStateChooseRace
			output.WriteString("Please choose a race from the following options:\r\n")

			/* Counter value for periodically line-breaking */
			index := 0

			for iter := Races.Head; iter != nil; iter = iter.Next {
				race := iter.Value.(*Race)

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

		client.ConnectionState = ConnectionStateChooseClass
		output.WriteString("\r\nPlease choose a job from the following options:\r\n")

		/* Counter value for periodically line-breaking */
		index := 0

		for iter := Jobs.Head; iter != nil; iter = iter.Next {
			job := iter.Value.(*Job)

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

		client.Character.Job = job
		client.ConnectionState = ConnectionStateConfirmClass
		output.WriteString(fmt.Sprintf("\r\nAre you sure you want to be a %s? [y/N] ", job.Name))

	case ConnectionStateConfirmClass:
		if !strings.HasPrefix(strings.ToLower(message), "y") {
			client.ConnectionState = ConnectionStateChooseClass
			output.WriteString("Please choose a job from the following options:\r\n")

			/* Counter value for periodically line-breaking */
			index := 0

			for iter := Jobs.Head; iter != nil; iter = iter.Next {
				job := iter.Value.(*Job)

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

		/* Start stat-rolling... */
		output.WriteString("\r\nOk.  Rolling stats for a new character...\r\n")
		output.WriteString("\r\nEach core attribute has a default range of 10-20.  A primary attribute\r\n")
		output.WriteString("increases the base and max by two; matching race and job bonuses stack.\r\n")

		output.WriteString(fmt.Sprintf("\r\nJob primary attribute:  %s\r\n", StatNameTable[client.Character.Job.PrimaryAttribute]))
		output.WriteString(fmt.Sprintf("Race primary attribute: %s\r\n\r\n", StatNameTable[client.Character.Race.PrimaryAttribute]))

		client.Character.RollStats()

		for index := STAT_NONE + 1; index < STAT_MAX; index++ {
			output.WriteString(fmt.Sprintf("%-15s %d\r\n", fmt.Sprintf("%s:", strings.ToTitle(StatNameTable[index])), client.Character.Stats[index]))
		}

		output.WriteString("\r\nAccept this stat roll? [y/N]")
		client.ConnectionState = ConnectionStateRollingStats

	case ConnectionStateRollingStats:
		if !strings.HasPrefix(strings.ToLower(message), "y") && client.remainingRolls >= 0 {
			client.ConnectionState = ConnectionStateRollingStats

			client.Character.RollStats()

			if client.remainingRolls == 0 {
				output.WriteString("\r\nFinal roll:\r\n")
			}

			for index := STAT_NONE + 1; index < STAT_MAX; index++ {
				output.WriteString(fmt.Sprintf("%-15s %d\r\n", fmt.Sprintf("%s:", strings.ToTitle(StatNameTable[index])), client.Character.Stats[index]))
			}

			if client.remainingRolls > 0 {
				output.WriteString("\r\nAccept this stat roll? [y/N]")
				output.WriteString(fmt.Sprintf("\r\n* After %d more rolls, you will be forced to accept.\r\n", client.remainingRolls))
			} else {
				output.WriteString("\r\n[ Press return to continue ]")
			}

			client.remainingRolls--
			break
		}

		err := client.Character.Finalize()
		if err != nil {
			log.Printf("Unable to create new character %v, dropping connection.\r\n", client.Character)
			client.conn.Close()
			break
		}

		output.WriteString(fmt.Sprintf("Final stats for new character %s:\r\n\r\n", client.Character.Name))
		for index := STAT_NONE + 1; index < STAT_MAX; index++ {
			output.WriteString(fmt.Sprintf("%-15s %d\r\n", fmt.Sprintf("%s:", strings.ToTitle(StatNameTable[index])), client.Character.Stats[index]))
		}

		output.WriteString("\r\n")

		client.ConnectionState = ConnectionStateMessageOfTheDay
		output.WriteString(string(Config.motd))
		output.WriteString("[ Press return to continue ]")

	case ConnectionStateMessageOfTheDay:
		client.ConnectionState = ConnectionStatePlaying

		game.Characters.Insert(client.Character)

		for iter := client.Character.Inventory.Head; iter != nil; iter = iter.Next {
			obj := iter.Value.(*ObjectInstance)

			game.Objects.Insert(obj)

			if obj.Contents != nil {
				for innerIter := obj.Contents.Head; innerIter != nil; innerIter = innerIter.Next {
					containedObj := innerIter.Value.(*ObjectInstance)

					game.Objects.Insert(containedObj)
				}
			}
		}

		if client.Character.Room != nil {
			client.Character.Room.AddCharacter(client.Character)

			out := fmt.Sprintf("{W%s has entered the game.{x\r\n", client.Character.Name)

			game.broadcast(out, func(character *Character) bool {
				return character != client.Character
			})
		}

		do_time(client.Character, "")
		client.Character.Send(fmt.Sprintf("%s\r\n", JoinedGameFlavourText))
		err := client.Character.syncJobSkills()
		if err != nil {
			log.Println(err)
		}

		do_look(client.Character, "")
	}

	if client.ConnectionState != ConnectionStatePlaying && output.Len() > 0 {
		client.Send(output.Bytes())
	}
}
