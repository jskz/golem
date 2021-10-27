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
	Id       uint
	ItemType string

	Name             string
	ShortDescription string
	LongDescription  string
	Description      string
	Flags            int

	Value0 int
	Value1 int
	Value2 int
	Value3 int
}

type ObjectInstance struct {
	Game      *Game           `json:"game"`
	Contents  *LinkedList     `json:"contents"`
	Inside    *ObjectInstance `json:"inside"`
	InRoom    *Room           `json:"inRoom"`
	CarriedBy *Character      `json:"carriedBy"`

	Id       uint   `json:"id"`
	ParentId uint   `json:"parentId"`
	ItemType string `json:"itemType"`

	Name             string `json:"name"`
	ShortDescription string `json:"shortDescription"`
	LongDescription  string `json:"longDescription"`
	Description      string `json:"description"`
	Flags            int    `json:"flags"`

	WearLocation int `json:"wearLocation"`

	Value0 int `json:"value0"`
	Value1 int `json:"value1"`
	Value2 int `json:"value2"`
	Value3 int `json:"value3"`

	CreatedAt time.Time `json:"createdAt"`
	Ttl       int       `json:"ttl"`
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
	ItemTypeTreasure  = "treasure"
	ItemTypeReagent   = "reagent"
	ItemTypeArtifact  = "artifact"
)

const (
	ITEM_TAKE           = 1
	ITEM_WEAPON         = 1 << 1
	ITEM_WEARABLE       = 1 << 2
	ITEM_DECAYS         = 1 << 3
	ITEM_DECAY_SILENTLY = 1 << 4
	ITEM_WEAR_HELD      = 1 << 5
	ITEM_WEAR_HEAD      = 1 << 6
	ITEM_WEAR_TORSO     = 1 << 7
	ITEM_WEAR_BODY      = 1 << 8
	ITEM_WEAR_NECK      = 1 << 9
	ITEM_WEAR_LEGS      = 1 << 10
	ITEM_WEAR_HANDS     = 1 << 11
	ITEM_WEAR_SHIELD    = 1 << 12
	ITEM_WEAR_ARMS      = 1 << 13
	ITEM_WEAR_WAIST     = 1 << 14
	ITEM_WEAR_FEET      = 1 << 15
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
	err := row.Scan(&obj.Id, &obj.Name, &obj.ShortDescription, &obj.LongDescription, &obj.Description, &obj.Flags, &obj.ItemType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return obj, nil
}

func (obj *ObjectInstance) GetShortDescription(viewer *Character) string {
	return obj.ShortDescription
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

	if obj.Contents != nil && obj.Contents.Count > 0 {
		for iter := obj.Contents.Head; iter != nil; iter = iter.Next {
			containedObject := iter.Value.(*ObjectInstance)
			containedObject.Finalize(obj)
		}
	}

	return nil
}

func (obj *ObjectInstance) Finalize(container *ObjectInstance) error {
	if obj == nil || obj.Id > 0 {
		return nil
	}

	var insideObjectInstanceId *uint = nil

	if container != nil {
		insideObjectInstanceId = &container.Id
	}

	result, err := obj.Game.db.Exec(`
		INSERT INTO
			object_instances(parent_id, inside_object_instance_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, obj.ParentId, insideObjectInstanceId, obj.Name, obj.ShortDescription, obj.LongDescription, obj.Description, obj.Flags, obj.ItemType, obj.Value0, obj.Value1, obj.Value2, obj.Value3)
	if err != nil {
		log.Printf("Failed to finalize new object: %v.\r\n", err)
		return err
	}

	objectInstanceId, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed to retrieve insert id: %v.\r\n", err)
		return err
	}

	obj.Id = uint(objectInstanceId)
	return nil
}

func (container *ObjectInstance) addObject(obj *ObjectInstance) {
	container.Contents.Insert(obj)

	obj.Inside = container
	obj.CarriedBy = nil
	obj.InRoom = nil
}

func (container *ObjectInstance) removeObject(obj *ObjectInstance) {
	container.Contents.Remove(obj)

	obj.Inside = nil
	obj.CarriedBy = nil
	obj.InRoom = nil
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
