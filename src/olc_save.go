/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "errors"

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
