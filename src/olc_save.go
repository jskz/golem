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
