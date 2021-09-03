/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

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
	ch.game = game

	var raceId uint
	var jobId uint

	err := row.Scan(&ch.id,
		&ch.name,
		&ch.shortDescription,
		&ch.longDescription,
		&ch.description,
		&raceId,
		&jobId,
		&ch.level,
		&ch.experience,
		&ch.health,
		&ch.maxHealth,
		&ch.mana,
		&ch.maxMana,
		&ch.stamina,
		&ch.maxStamina,
		&ch.strength,
		&ch.dexterity,
		&ch.intelligence,
		&ch.wisdom,
		&ch.constitution,
		&ch.charisma,
		&ch.luck)
	if err != nil {
		return nil, err
	}

	/* Sanity check for pointers by race and job id before continuing */
	_, ok := RaceTable[raceId]
	if !ok {
		return nil, nil
	}

	_, ok = JobTable[jobId]
	if !ok {
		return nil, nil
	}

	ch.race = RaceTable[raceId]
	ch.job = JobTable[jobId]

	return ch, nil
}
