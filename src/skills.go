/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

type Proficiency struct {
	id          int
	skillId     int
	proficiency int
	level       int
	complexity  int
	cost        int
}

func do_skills(ch *Character, arguments string) {
	ch.Send("Not yet implemented, try again soon!\r\n")
}

func do_practice(ch *Character, arguments string) {
	ch.Send("Not yet implemented, try again soon!\r\n")
}

func (ch *Character) LoadPlayerSkills() error {
	rows, err := ch.game.db.Query(`
		SELECT
			pc_skill_proficiency.id,
			pc_skill_proficiency.skill_id,
			pc_skill_proficiency.proficiency,

			job_skill.level,
			job_skill.complexity,
			job_skill.cost
		FROM
			pc_skill_proficiency
		INNER JOIN
			job_skill
		ON
			job_skill.id = pc_skill_proficiency.player_character_id
		WHERE
			pc_skill_proficiency.player_character_id = ?
	`, ch.id)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		proficiency := &Proficiency{}

		err := rows.Scan(&proficiency.id, &proficiency.skillId, &proficiency.proficiency, &proficiency.level, &proficiency.complexity, &proficiency.cost)
		if err != nil {
			return err
		}

		ch.skills[proficiency.id] = proficiency
	}

	return nil
}
