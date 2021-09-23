/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"database/sql"
	"fmt"
	"strings"
	"unicode"
)

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
	contents  *LinkedList
	inside    *ObjectInstance
	inRoom    *Room
	carriedBy *Character

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

const (
	ItemTypeNone      = "protoplasm"
	ItemTypeContainer = "container"
	ItemTypeArmor     = "armor"
	ItemTypeWeapon    = "weapon"
	ItemTypeLight     = "light"
	ItemTypeFurniture = "furniture"
)

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

func (obj *ObjectInstance) getShortDescription(viewer *Character) string {
	return obj.shortDescription
}

func (obj *ObjectInstance) getShortDescriptionUpper(viewer *Character) string {
	var short string = obj.getShortDescription(viewer)

	if short == "" {
		return ""
	}

	runes := []rune(short)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (container *ObjectInstance) addObject(obj *ObjectInstance) {
	container.contents.Insert(obj)

	obj.inside = container
	obj.carriedBy = nil
	obj.inRoom = nil
}

func (ch *Character) showObjectList(objects *LinkedList) {
	var output strings.Builder

	if objects == nil || objects.Count < 1 {
		return
	}

	for iter := objects.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		output.WriteString(fmt.Sprintf("%s\r\n", obj.carriedBy.getShortDescriptionUpper(ch)))
	}

	ch.Send(output.String())
}
