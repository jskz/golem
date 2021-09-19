/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "database/sql"

type Object struct {
	id       uint
	itemType string

	name             string
	shortDescription string
	longDescription  string
	description      string

	value0 int
	value1 int
	value2 int
	value3 int
}

type ObjectInstance struct {
	id       uint
	parentId uint
	itemType string

	name             string
	shortDescription string
	longDescription  string
	description      string

	value0 int
	value1 int
	value2 int
	value3 int
}

func (game *Game) LoadObjectIndex(index uint) (*Object, error) {
	row := game.db.QueryRow(`
		SELECT
			id,
			name,
			short_description,
			long_description,
			description,
			item_type
		FROM
			objects
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`, index)

	obj := &Object{}
	err := row.Scan(&obj.id, &obj.name, &obj.shortDescription, &obj.longDescription, &obj.description, &obj.itemType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return obj, nil
}
