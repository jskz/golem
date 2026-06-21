/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

type Social struct {
	Id           uint
	Name         string
	CharNoArg    string
	OthersNoArg  string
	CharFound    string
	OthersFound  string
	VictFound    string
	CharNotFound string
	CharAuto     string
	OthersAuto   string
}

func (game *Game) LoadSocials() error {
	log.Printf("Loading socials.\r\n")

	game.socials = make(map[string]*Social)

	rows, err := game.db.Query(`
		SELECT
			id,
			name,
			char_no_arg,
			others_no_arg,
			char_found,
			others_found,
			vict_found,
			char_not_found,
			char_auto,
			others_auto
		FROM
			socials
		WHERE
			deleted_at IS NULL
		ORDER BY
			name
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		social := &Social{}

		err := rows.Scan(
			&social.Id,
			&social.Name,
			&social.CharNoArg,
			&social.OthersNoArg,
			&social.CharFound,
			&social.OthersFound,
			&social.VictFound,
			&social.CharNotFound,
			&social.CharAuto,
			&social.OthersAuto,
		)
		if err != nil {
			return err
		}

		social.Name = strings.ToLower(social.Name)
		game.socials[social.Name] = social
	}

	if err := rows.Err(); err != nil {
		return err
	}

	log.Printf("Loaded %d socials from database.\r\n", len(game.socials))
	return nil
}

func (game *Game) FindSocialByName(name string) *Social {
	if game == nil || game.socials == nil {
		return nil
	}

	return game.socials[strings.ToLower(name)]
}

func (ch *Character) trySocial(name string, arguments string) bool {
	if ch.Game == nil {
		return false
	}

	if ch.Game.FindSocialByName(name) == nil {
		return false
	}

	if !ch.canActFromPosition(PositionResting) {
		return true
	}

	do_social(ch, name, arguments)
	return true
}

func do_social(ch *Character, name string, arguments string) {
	social := ch.Game.FindSocialByName(name)
	if social == nil {
		ch.Send("{RAlas, there is no such social.{x\r\n")
		return
	}

	arg, _ := OneArgument(arguments)
	if arg == "" {
		ch.Send(formatSocialMessage(social.CharNoArg, ch, nil, ch))
		sendToRoomExcept(ch, func(rch *Character) string {
			return formatSocialMessage(social.OthersNoArg, ch, nil, rch)
		})
		return
	}

	victim := ch.FindCharacterInRoom(arg)
	if victim == nil {
		ch.Send(social.CharNotFound + "\r\n")
		return
	}

	if victim == ch {
		ch.Send(formatSocialMessage(social.CharAuto, ch, victim, ch))
		sendToRoomExcept(ch, func(rch *Character) string {
			return formatSocialMessage(social.OthersAuto, ch, victim, rch)
		})
		return
	}

	sendToRoomExcept(ch, func(rch *Character) string {
		if rch == victim {
			return formatSocialMessage(social.VictFound, ch, victim, rch)
		}

		return formatSocialMessage(social.OthersFound, ch, victim, rch)
	})

	ch.Send(formatSocialMessage(social.CharFound, ch, victim, ch))
}

func do_socials(ch *Character, arguments string) {
	if ch.Game == nil {
		ch.Send("There are no socials available.\r\n")
		return
	}

	names := make([]string, 0, len(ch.Game.socials))
	for name := range ch.Game.socials {
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		ch.Send("There are no socials available.\r\n")
		return
	}

	var output strings.Builder
	for i, name := range names {
		output.WriteString(fmt.Sprintf("%-12s", name))
		if (i+1)%6 == 0 {
			output.WriteString("\r\n")
		}
	}

	if len(names)%6 != 0 {
		output.WriteString("\r\n")
	}

	ch.Send(output.String())
}

func formatSocialMessage(template string, actor *Character, victim *Character, viewer *Character) string {
	if actor == nil {
		return ""
	}

	if template == "" {
		return ""
	}

	output := template
	output = strings.ReplaceAll(output, "$n", actor.GetShortDescriptionUpper(viewer)+"{x")
	output = strings.ReplaceAll(output, "$N", victimShortDescription(victim, viewer)+"{x")

	return fmt.Sprintf("%s\r\n", output)
}

func victimShortDescription(victim *Character, viewer *Character) string {
	if victim == nil {
		return ""
	}

	return victim.GetShortDescription(viewer)
}
