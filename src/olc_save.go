/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"errors"
	"log"
)

func (room *Room) Save() error {
	if room.Flags&ROOM_VIRTUAL != 0 || room.Flags&ROOM_PLANAR != 0 {
		return errors.New("attempt to save a virtual room")
	}

	_, err := room.Game.db.Exec(`
		UPDATE
			rooms
		SET
			name = ?,
			description = ?,
			flags = ?
		WHERE
			id = ?
	`, room.Name, room.Description, room.Flags, room.Id)
	if err != nil {
		return err
	}

	return nil
}

func (zone *Zone) Save() error {
	_, err := zone.Game.db.Exec(`
		UPDATE
			zones
		SET
			name = ?,
			who_description = ?,
			low = ?,
			high = ?,
			reset_message = ?,
			reset_frequency = ?
		WHERE
			id = ?
	`, zone.Name, zone.WhoDescription, zone.Low, zone.High, zone.ResetMessage, zone.ResetFrequency, zone.Id)
	if err != nil {
		return err
	}

	return nil
}

// Updates an object's parent index with this instance's properties
func (obj *ObjectInstance) Sync() error {
	if obj.ParentId == 0 {
		return errors.New("trying to sync an object instance with no parent ID")
	}

	_, err := obj.Game.db.Exec(`
		UPDATE
			objects
		SET
			name = ?,
			short_description = ?,
			long_description = ?,
			description = ?,
			flags = ?,
			item_type = ?,
			value_1 = ?,
			value_2 = ?,
			value_3 = ?,
			value_4 = ?
		WHERE
			id = ?
	`, obj.Name, obj.ShortDescription, obj.LongDescription, obj.Description, obj.Flags, obj.ItemType, obj.Value0, obj.Value1, obj.Value2, obj.Value3, obj.ParentId)
	if err != nil {
		return err
	}

	return nil
}

// Updates a mob instance's index with this instance's properties
func (ch *Character) Sync() error {
	if ch.Flags&CHAR_IS_PLAYER != 0 || ch.Id == 0 {
		return errors.New("invalid NPC tried to sync")
	}

	_, err := ch.Game.db.Exec(`
		UPDATE
			mobiles
		SET
			name = ?,
			short_description = ?,
			long_description = ?,
			description = ?,
			race_id = ?,
			job_id = ?,
			flags = ?,
			level = ?,
			gold = ?,
			experience = ?,
			health = ?,
			max_health = ?,
			mana = ?,
			max_mana = ?,
			stamina = ?,
			max_stamina = ?,
			stat_str = ?,
			stat_dex = ?,
			stat_int = ?,
			stat_wis = ?,
			stat_con = ?,
			stat_cha = ?,
			stat_lck = ?
		WHERE
			id = ?
	`,
		ch.Name,
		ch.ShortDescription,
		ch.LongDescription,
		ch.Description,
		ch.Race.Id,
		ch.Job.Id,
		ch.Flags,
		ch.Level,
		ch.Gold,
		ch.Experience,
		ch.Health,
		ch.MaxHealth,
		ch.Mana,
		ch.MaxMana,
		ch.Stamina,
		ch.MaxStamina,
		ch.Stats[STAT_STRENGTH],
		ch.Stats[STAT_DEXTERITY],
		ch.Stats[STAT_INTELLIGENCE],
		ch.Stats[STAT_WISDOM],
		ch.Stats[STAT_CONSTITUTION],
		ch.Stats[STAT_CHARISMA],
		ch.Stats[STAT_LUCK],
		ch.Id)
	if err != nil {
		return err
	}

	return nil
}

func (reset *Reset) Delete() error {
	_, err := reset.Zone.Game.db.Exec(`
		DELETE FROM
			resets
		WHERE
			id = ?
	`, reset.Id)
	if err != nil {
		return err
	}

	return nil
}

func (exit *Exit) Delete() error {
	_, err := exit.Room.Game.db.Exec(`
		DELETE FROM
			exits
		WHERE
			id = ?
	`, exit.Id)
	if err != nil {
		return err
	}

	return nil
}

func (exit *Exit) Finalize() error {
	if exit.Id != 0 {
		return errors.New("trying to finalize an exit when already finalized")
	}

	if exit.Room.Id == 0 || exit.To.Id == 0 {
		return errors.New("currently unsupported saving exit between virtual rooms")
	}

	result, err := exit.Room.Game.db.Exec(`
		INSERT INTO
			exits(room_id, to_room_id, direction, flags)
		VALUES
			(?, ?, ?, ?)
	`, exit.Room.Id, exit.To.Id, exit.Direction, exit.Flags)
	if err != nil {
		log.Printf("Failed to finalize exit: %v.\r\n", err)
		return err
	}

	exitId, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve insert id: %v.\r\n", err)
		return err
	}

	exit.Id = uint(exitId)
	return nil
}

func (exit *Exit) Save() error {
	if exit.Id == 0 {
		return errors.New("trying to update an exit before it was finalized")
	}

	_, err := exit.Room.Game.db.Exec(`
		UPDATE
			exits
		SET
			flags = ?
		WHERE
			id = ?
	`, exit.Flags, exit.Id)
	if err != nil {
		return err
	}

	return nil
}
