/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

func (room *Room) Save() error {
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
