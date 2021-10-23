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
	"log"
	"strings"
	"time"
	"unicode"
)

type Object struct {
	id       uint
	itemType string

	name             string
	shortDescription string
	longDescription  string
	description      string
	flags            int

	value0 int
	value1 int
	value2 int
	value3 int
}

type ObjectInstance struct {
	game      *Game           `json:"game"`
	contents  *LinkedList     `json:"contents"`
	inside    *ObjectInstance `json:"inside"`
	inRoom    *Room           `json:"inRoom"`
	carriedBy *Character      `json:"carriedBy"`

	id       uint   `json:"id"`
	parentId uint   `json:"parentId"`
	itemType string `json:"itemType"`

	name             string `json:"name"`
	shortDescription string `json:"shortDescription"`
	longDescription  string `json:"longDescription"`
	description      string `json:"description"`
	flags            int    `json:"flags"`

	WearLocation int `json:"wearLocation"`

	value0 int `json:"value0"`
	value1 int `json:"value1"`
	value2 int `json:"value2"`
	value3 int `json:"value3"`

	createdAt time.Time `json:"createdAt"`
	ttl       int       `json:"ttl"`
}

const (
	ItemTypeNone      = "protoplasm"
	ItemTypeContainer = "container"
	ItemTypeScroll    = "scroll"
	ItemTypePotion    = "potion"
	ItemTypeArmor     = "armor"
	ItemTypeWeapon    = "weapon"
	ItemTypeLight     = "light"
	ItemTypeFurniture = "furniture"
	ItemTypeSign      = "sign"
)

const (
	ITEM_TAKE           = 1
	ITEM_WEAPON         = 1 << 1
	ITEM_WEARABLE       = 1 << 2
	ITEM_DECAYS         = 1 << 3
	ITEM_DECAY_SILENTLY = 1 << 4
)

func (game *Game) LoadObjectIndex(index uint) (*Object, error) {
	row := game.db.QueryRow(`
		SELECT
			id,
			name,
			short_description,
			long_description,
			description,
			flags,
			item_type
		FROM
			objects
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`, index)

	obj := &Object{}
	err := row.Scan(&obj.id, &obj.name, &obj.shortDescription, &obj.longDescription, &obj.description, &obj.flags, &obj.itemType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return obj, nil
}

func (obj *ObjectInstance) GetShortDescription(viewer *Character) string {
	return obj.shortDescription
}

func (obj *ObjectInstance) GetShortDescriptionUpper(viewer *Character) string {
	var short string = obj.GetShortDescription(viewer)

	if short == "" {
		return ""
	}

	runes := []rune(short)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (obj *ObjectInstance) reify() error {
	obj.Finalize(nil)

	if obj.contents != nil && obj.contents.Count > 0 {
		for iter := obj.contents.Head; iter != nil; iter = iter.Next {
			containedObject := iter.Value.(*ObjectInstance)
			containedObject.Finalize(obj)
		}
	}

	return nil
}

func (obj *ObjectInstance) Finalize(container *ObjectInstance) error {
	if obj == nil || obj.id > 0 {
		return nil
	}

	var insideObjectInstanceId *uint = nil

	if container != nil {
		insideObjectInstanceId = &container.id
	}

	result, err := obj.game.db.Exec(`
		INSERT INTO
			object_instances(parent_id, inside_object_instance_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, obj.parentId, insideObjectInstanceId, obj.name, obj.shortDescription, obj.longDescription, obj.description, obj.flags, obj.itemType, obj.value0, obj.value1, obj.value2, obj.value3)
	if err != nil {
		log.Printf("Failed to finalize new object: %v.\r\n", err)
		return err
	}

	objectInstanceId, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve insert id: %v.\r\n", err)
		return err
	}

	obj.id = uint(objectInstanceId)
	return nil
}

func (container *ObjectInstance) addObject(obj *ObjectInstance) {
	container.contents.Insert(obj)

	obj.inside = container
	obj.carriedBy = nil
	obj.inRoom = nil
}

func (container *ObjectInstance) removeObject(obj *ObjectInstance) {
	container.contents.Remove(obj)

	obj.inside = nil
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

		output.WriteString(fmt.Sprintf("  %s\r\n", obj.GetShortDescriptionUpper(ch)))
	}

	ch.Send(output.String())
}
