/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "errors"

func (game *Game) LoadMobileIndex(index uint) (*Character, error) {
	/* There was no online player with this name, search the database. */
	row := game.db.QueryRow(`
		SELECT
			id,
			name,
			short_description,
			long_description,
			description,
			race_id,
			job_id,
			flags,
			level,
			experience,
			health,
			max_health,
			mana,
			max_mana,
			stamina,
			max_stamina,
			stat_str,
			stat_dex,
			stat_int,
			stat_wis,
			stat_con,
			stat_cha,
			stat_lck
		FROM
			mobiles
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`, index)

	ch := NewCharacter()
	ch.Game = game

	var raceId uint
	var jobId uint

	err := row.Scan(&ch.Id,
		&ch.Name,
		&ch.ShortDescription,
		&ch.LongDescription,
		&ch.Description,
		&raceId,
		&jobId,
		&ch.Flags,
		&ch.Level,
		&ch.Experience,
		&ch.Health,
		&ch.MaxHealth,
		&ch.Mana,
		&ch.MaxMana,
		&ch.Stamina,
		&ch.MaxStamina,
		&ch.Strength,
		&ch.Dexterity,
		&ch.Intelligence,
		&ch.Wisdom,
		&ch.Constitution,
		&ch.Charisma,
		&ch.Luck)
	if err != nil {
		return nil, err
	}

	ch.Race = FindRaceByID(jobId)
	if ch.Race == nil {
		return nil, errors.New("failed to load race")
	}

	ch.Job = FindJobByID(jobId)
	if ch.Job == nil {
		return nil, errors.New("failed to load job")
	}

	return ch, nil
}
