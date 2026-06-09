/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"context"
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
	Weight float64
	Ttl    int
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

	Value0 int     `json:"value0"`
	Value1 int     `json:"value1"`
	Value2 int     `json:"value2"`
	Value3 int     `json:"value3"`
	Weight float64 `json:"weight"`

	CreatedAt time.Time `json:"createdAt"`
	Ttl       int       `json:"ttl"`
}

const (
	ItemTypeNone           = "protoplasm"
	ItemTypeContainer      = "container"
	ItemTypeScroll         = "scroll"
	ItemTypePotion         = "potion"
	ItemTypeFood           = "food"
	ItemTypeDrinkContainer = "drink_container"
	ItemTypeFountain       = "fountain"
	ItemTypeArmor          = "armor"
	ItemTypeWeapon         = "weapon"
	ItemTypeLight          = "light"
	ItemTypeFurniture      = "furniture"
	ItemTypeSign           = "sign"
	ItemTypeTreasure       = "treasure"
	ItemTypeReagent        = "reagent"
	ItemTypeArtifact       = "artifact"
	ItemTypeCurrency       = "currency"
)

const (
	FURNITURE_STAND_AT = 1
	FURNITURE_STAND_ON = 1 << 1
	FURNITURE_STAND_IN = 1 << 2
	FURNITURE_SIT_AT   = 1 << 3
	FURNITURE_SIT_ON   = 1 << 4
	FURNITURE_SIT_IN   = 1 << 5
	FURNITURE_REST_AT  = 1 << 6
	FURNITURE_REST_ON  = 1 << 7
	FURNITURE_REST_IN  = 1 << 8
	FURNITURE_SLEEP_AT = 1 << 9
	FURNITURE_SLEEP_ON = 1 << 10
	FURNITURE_SLEEP_IN = 1 << 11
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
	ITEM_GLOW           = 1 << 16
	ITEM_HUM            = 1 << 17
	ITEM_CLOSED         = 1 << 18
	ITEM_CLOSEABLE      = 1 << 19
	ITEM_LOCKED         = 1 << 20
	ITEM_PERSISTENT     = 1 << 21
)

const ObjectGoldSingle = 2
const ObjectGoldCoins = 3
const DefaultObjectDecayTtl = 20

type ObjectFlag struct {
	Name string `json:"name"`
	Flag int    `json:"flag"`
}

var ObjectFlagTable []ObjectFlag = []ObjectFlag{
	{Name: "item", Flag: ITEM_TAKE},
	{Name: "weapon", Flag: ITEM_WEAPON},
	{Name: "wearable", Flag: ITEM_WEARABLE},
	{Name: "decays", Flag: ITEM_DECAYS},
	{Name: "decay_silently", Flag: ITEM_DECAY_SILENTLY},
	{Name: "wear_held", Flag: ITEM_WEAR_HELD},
	{Name: "wear_head", Flag: ITEM_WEAR_HEAD},
	{Name: "wear_torso", Flag: ITEM_WEAR_TORSO},
	{Name: "wear_body", Flag: ITEM_WEAR_BODY},
	{Name: "wear_neck", Flag: ITEM_WEAR_NECK},
	{Name: "wear_legs", Flag: ITEM_WEAR_LEGS},
	{Name: "wear_hands", Flag: ITEM_WEAR_HANDS},
	{Name: "wear_shield", Flag: ITEM_WEAR_SHIELD},
	{Name: "wear_arms", Flag: ITEM_WEAR_ARMS},
	{Name: "wear_waist", Flag: ITEM_WEAR_WAIST},
	{Name: "wear_feet", Flag: ITEM_WEAR_FEET},
	{Name: "glow", Flag: ITEM_GLOW},
	{Name: "hum", Flag: ITEM_HUM},
	{Name: "closed", Flag: ITEM_CLOSED},
	{Name: "closeable", Flag: ITEM_CLOSEABLE},
	{Name: "locked", Flag: ITEM_LOCKED},
	{Name: "persistent", Flag: ITEM_PERSISTENT},
}

var FurnitureFlagTable []Flag = []Flag{
	{Name: "stand_at", Flag: FURNITURE_STAND_AT},
	{Name: "stand_on", Flag: FURNITURE_STAND_ON},
	{Name: "stand_in", Flag: FURNITURE_STAND_IN},
	{Name: "sit_at", Flag: FURNITURE_SIT_AT},
	{Name: "sit_on", Flag: FURNITURE_SIT_ON},
	{Name: "sit_in", Flag: FURNITURE_SIT_IN},
	{Name: "rest_at", Flag: FURNITURE_REST_AT},
	{Name: "rest_on", Flag: FURNITURE_REST_ON},
	{Name: "rest_in", Flag: FURNITURE_REST_IN},
	{Name: "sleep_at", Flag: FURNITURE_SLEEP_AT},
	{Name: "sleep_on", Flag: FURNITURE_SLEEP_ON},
	{Name: "sleep_in", Flag: FURNITURE_SLEEP_IN},
}

func FindObjectFlag(flag string) *ObjectFlag {
	for _, f := range ObjectFlagTable {
		if strings.EqualFold(f.Name, flag) {
			return &f
		}
	}

	return nil
}

func FindFurnitureFlag(flag string) *Flag {
	for _, f := range FurnitureFlagTable {
		if strings.EqualFold(f.Name, flag) {
			return &f
		}
	}

	return nil
}

func normalizeObjectTtl(flags int, ttl int) int {
	if ttl < 0 {
		ttl = 0
	}

	if flags&ITEM_DECAYS != 0 && ttl == 0 {
		return DefaultObjectDecayTtl
	}

	return ttl
}

func objectCreatedAtFromUnix(createdAt sql.NullInt64) time.Time {
	if createdAt.Valid && createdAt.Int64 > 0 {
		return time.Unix(createdAt.Int64, 0)
	}

	return time.Now()
}

func (obj *ObjectInstance) ensureDecayState(now time.Time) {
	if obj == nil {
		return
	}

	ttlUnset := obj.Ttl <= 0
	if obj.CreatedAt.IsZero() || (obj.Flags&ITEM_DECAYS != 0 && ttlUnset) {
		obj.CreatedAt = now
	}

	obj.Ttl = normalizeObjectTtl(obj.Flags, obj.Ttl)
}

func (obj *ObjectInstance) StartDecay() {
	if obj == nil {
		return
	}

	obj.CreatedAt = time.Now()
	obj.Ttl = normalizeObjectTtl(obj.Flags, obj.Ttl)
}

func (obj *ObjectInstance) shouldDecay(now time.Time) bool {
	if obj == nil || obj.Flags&ITEM_DECAYS == 0 {
		return false
	}

	obj.ensureDecayState(now)

	if obj.Ttl <= 0 {
		return false
	}

	return int(now.Sub(obj.CreatedAt).Minutes()) >= obj.Ttl
}

func (game *Game) objectInstanceFromIndex(obj *Object) *ObjectInstance {
	if obj == nil {
		return nil
	}

	objectInstance := &ObjectInstance{
		Game:             game,
		ParentId:         obj.Id,
		Contents:         NewLinkedList(),
		Description:      obj.Description,
		ShortDescription: obj.ShortDescription,
		LongDescription:  obj.LongDescription,
		Name:             obj.Name,
		WearLocation:     -1,
		ItemType:         obj.ItemType,
		CreatedAt:        time.Now(),
		Flags:            obj.Flags,
		Value0:           obj.Value0,
		Value1:           obj.Value1,
		Value2:           obj.Value2,
		Value3:           obj.Value3,
		Weight:           normalizeObjectWeight(obj.Weight),
		Ttl:              normalizeObjectTtl(obj.Flags, obj.Ttl),
	}

	return objectInstance
}

func (game *Game) NewObjectInstance(objectIndex uint) *ObjectInstance {
	obj, err := game.LoadObjectIndex(objectIndex)
	if err != nil {
		log.Printf("Failed to create object instance from id %d: %v\r\n", objectIndex, err)
		return nil
	}

	if obj == nil {
		log.Printf("Failed to create object instance from id %d: object index does not exist\r\n", objectIndex)
		return nil
	}

	return game.objectInstanceFromIndex(obj)
}

func (game *Game) CreateGold(amount int) *ObjectInstance {
	if amount <= 0 {
		return nil
	} else if amount == 1 {
		obj, err := game.LoadObjectIndex(ObjectGoldSingle)
		if err != nil {
			log.Printf("Failed to create single gold coin: %v\r\n", err)
			return nil
		}

		objectInstance := &ObjectInstance{
			Game:             game,
			ParentId:         obj.Id,
			Description:      obj.Description,
			ShortDescription: obj.ShortDescription,
			LongDescription:  obj.LongDescription,
			Name:             obj.Name,
			ItemType:         obj.ItemType,
			CreatedAt:        time.Now(),
			Flags:            ITEM_TAKE,
			WearLocation:     0,
			Value0:           amount,
			Weight:           normalizeObjectWeight(obj.Weight),
		}

		return objectInstance
	}

	obj, err := game.LoadObjectIndex(ObjectGoldCoins)
	if err != nil {
		log.Printf("Failed to create gold coins: %v\r\n", err)
		return nil
	}

	objectInstance := &ObjectInstance{
		Game:             game,
		ParentId:         obj.Id,
		ShortDescription: fmt.Sprintf(obj.ShortDescription, amount),
		LongDescription:  obj.LongDescription,
		Description:      fmt.Sprintf(obj.Description, amount),
		Name:             obj.Name,
		ItemType:         obj.ItemType,
		CreatedAt:        time.Now(),
		Flags:            ITEM_TAKE,
		WearLocation:     0,
		Value0:           amount,
		Weight:           normalizeObjectWeight(obj.Weight),
	}

	return objectInstance
}

func (obj *ObjectInstance) GetFlagsString() string {
	var buf strings.Builder

	for _, flag := range ObjectFlagTable {
		if obj.Flags&flag.Flag != 0 {
			buf.WriteString(fmt.Sprintf("%s ", flag.Name))
		}
	}

	if obj.Flags == 0 {
		buf.WriteString("none")
	}

	return strings.TrimRight(buf.String(), " ")
}

func (obj *ObjectInstance) GetFurnitureFlagsString() string {
	var buf strings.Builder

	if obj == nil || obj.Value2 == 0 {
		return "none"
	}

	for _, flag := range FurnitureFlagTable {
		if obj.Value2&flag.Flag != 0 {
			buf.WriteString(fmt.Sprintf("%s ", flag.Name))
		}
	}

	if buf.Len() == 0 {
		return "none"
	}

	return strings.TrimRight(buf.String(), " ")
}

func furnitureFlagsForPosition(position int) (int, int, int) {
	switch position {
	case PositionStanding:
		return FURNITURE_STAND_AT, FURNITURE_STAND_ON, FURNITURE_STAND_IN
	case PositionSitting:
		return FURNITURE_SIT_AT, FURNITURE_SIT_ON, FURNITURE_SIT_IN
	case PositionResting:
		return FURNITURE_REST_AT, FURNITURE_REST_ON, FURNITURE_REST_IN
	case PositionSleeping:
		return FURNITURE_SLEEP_AT, FURNITURE_SLEEP_ON, FURNITURE_SLEEP_IN
	default:
		return 0, 0, 0
	}
}

func (obj *ObjectInstance) FurnitureRelation(position int) (string, bool) {
	if obj == nil || obj.ItemType != ItemTypeFurniture {
		return "", false
	}

	at, on, in := furnitureFlagsForPosition(position)
	switch {
	case at != 0 && obj.Value2&at != 0:
		return "at", true
	case on != 0 && obj.Value2&on != 0:
		return "on", true
	case in != 0 && obj.Value2&in != 0:
		return "in", true
	default:
		return "", false
	}
}

func (obj *ObjectInstance) SupportsFurniturePosition(position int) bool {
	_, ok := obj.FurnitureRelation(position)
	return ok
}

func (obj *ObjectInstance) CountFurnitureUsers() int {
	if obj == nil || obj.InRoom == nil || obj.InRoom.Characters == nil {
		return 0
	}

	count := 0
	for iter := obj.InRoom.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)
		if rch.Furniture == obj {
			count++
		}
	}

	return count
}

func (game *Game) LoadObjectsByIndices(indices []uint) ([]*Object, error) {
	var objectIdValues strings.Builder

	if len(indices) == 0 {
		return make([]*Object, 0), nil
	}

	for _, index := range indices {
		objectIdValues.WriteString(fmt.Sprintf("%d,", index))
	}

	objectIdValuesString := strings.TrimRight(objectIdValues.String(), ",")

	rows, err := game.db.Query(fmt.Sprintf(`
		SELECT
			id,
			name,
			short_description,
			long_description,
			description,
			flags,
			item_type,
			value_1,
			value_2,
			value_3,
			value_4,
			weight,
			ttl
		FROM
			objects
		WHERE
			id 
		IN
			(%s)
		AND
			deleted_at IS NULL
	`, objectIdValuesString))
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var objects []*Object = make([]*Object, 0)

	for rows.Next() {
		obj := &Object{}
		err := rows.Scan(&obj.Id, &obj.Name, &obj.ShortDescription, &obj.LongDescription, &obj.Description, &obj.Flags, &obj.ItemType, &obj.Value0, &obj.Value1, &obj.Value2, &obj.Value3, &obj.Weight, &obj.Ttl)

		if err != nil {
			if err == sql.ErrNoRows {
				break
			}

			return nil, err
		}

		objects = append(objects, obj)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return objects, nil
}

func (game *Game) LoadObjectIndex(index uint) (*Object, error) {
	row := game.db.QueryRow(`
		SELECT
			id,
			name,
			short_description,
			long_description,
			description,
			flags,
			item_type,
			value_1,
			value_2,
			value_3,
			value_4,
			weight,
			ttl
		FROM
			objects
		WHERE
			id = ?
		AND
			deleted_at IS NULL
	`, index)

	obj := &Object{}
	err := row.Scan(&obj.Id, &obj.Name, &obj.ShortDescription, &obj.LongDescription, &obj.Description, &obj.Flags, &obj.ItemType, &obj.Value0, &obj.Value1, &obj.Value2, &obj.Value3, &obj.Weight, &obj.Ttl)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return obj, nil
}

func (obj *ObjectInstance) GetShortDescription(viewer *Character) string {
	if !obj.Visible(viewer) {
		return "something"
	}

	return obj.ShortDescription
}

func (obj *ObjectInstance) GetShortDescriptionUpper(viewer *Character) string {
	if !obj.Visible(viewer) {
		return "Something"
	}

	var short string = obj.GetShortDescription(viewer)

	if short == "" {
		return ""
	}

	runes := []rune(short)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

func (obj *ObjectInstance) Visible(viewer *Character) bool {
	if viewer.Affected&AFFECT_BLINDNESS != 0 {
		return false
	}

	if viewer.Room != nil && viewer.Room.Flags&ROOM_DARK != 0 && !viewer.Room.ActiveLightSourcePresent() {
		return false
	}

	return true
}

func normalizeObjectWeight(weight float64) float64 {
	if weight < 0 {
		return 0
	}

	return weight
}

func (obj *ObjectInstance) GetWeight() float64 {
	if obj == nil {
		return 0
	}

	return normalizeObjectWeight(obj.Weight)
}

func (obj *ObjectInstance) GetContentsWeight() float64 {
	if obj == nil || obj.Contents == nil {
		return 0
	}

	var total float64
	for iter := obj.Contents.Head; iter != nil; iter = iter.Next {
		containedObject := iter.Value.(*ObjectInstance)
		total += containedObject.GetTotalWeight()
	}

	return total
}

func (obj *ObjectInstance) GetTotalWeight() float64 {
	if obj == nil {
		return 0
	}

	return obj.GetWeight() + obj.GetContentsWeight()
}

func (obj *ObjectInstance) reifyTx(ctx context.Context, tx *sql.Tx) ([]*ObjectInstance, error) {
	return obj.reifyInContainerTx(ctx, tx, nil)
}

func (obj *ObjectInstance) reifyInContainerTx(ctx context.Context, tx *sql.Tx, container *ObjectInstance) ([]*ObjectInstance, error) {
	if obj == nil {
		return nil, nil
	}

	reified := make([]*ObjectInstance, 0)

	if obj.Id == 0 {
		obj.ensureDecayState(time.Now())

		var insideObjectInstanceId *uint = nil
		if container != nil {
			insideObjectInstanceId = &container.Id
		}

		result, err := tx.ExecContext(ctx, `
			INSERT INTO
				object_instances(parent_id, inside_object_instance_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4, weight, ttl, created_at)
			VALUES
				(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, obj.ParentId, insideObjectInstanceId, obj.Name, obj.ShortDescription, obj.LongDescription, obj.Description, obj.Flags, obj.ItemType, obj.Value0, obj.Value1, obj.Value2, obj.Value3, obj.GetWeight(), obj.Ttl, obj.CreatedAt)
		if err != nil {
			log.Printf("Failed to finalize new object: %v.\r\n", err)
			return reified, err
		}

		objectInstanceId, err := result.LastInsertId()
		if err != nil {
			log.Printf("Failed to retrieve insert id: %v.\r\n", err)
			return reified, err
		}

		obj.Id = uint(objectInstanceId)
		reified = append(reified, obj)
	}

	if obj.Contents != nil && obj.Contents.Count > 0 {
		for iter := obj.Contents.Head; iter != nil; iter = iter.Next {
			containedObject := iter.Value.(*ObjectInstance)

			containedReified, err := containedObject.reifyInContainerTx(ctx, tx, obj)
			reified = append(reified, containedReified...)
			if err != nil {
				return reified, err
			}
		}
	}

	return reified, nil
}

func (obj *ObjectInstance) objectInstanceIDs() []uint {
	if obj == nil {
		return nil
	}

	ids := make([]uint, 0)
	if obj.Id > 0 {
		ids = append(ids, obj.Id)
	}

	if obj.Contents != nil {
		for iter := obj.Contents.Head; iter != nil; iter = iter.Next {
			containedObject := iter.Value.(*ObjectInstance)
			ids = append(ids, containedObject.objectInstanceIDs()...)
		}
	}

	return ids
}

func (obj *ObjectInstance) resetObjectInstanceIDs() {
	if obj == nil {
		return
	}

	obj.Id = 0

	if obj.Contents != nil {
		for iter := obj.Contents.Head; iter != nil; iter = iter.Next {
			containedObject := iter.Value.(*ObjectInstance)
			containedObject.resetObjectInstanceIDs()
		}
	}
}

func (obj *ObjectInstance) persistedOwner() *Character {
	for current := obj; current != nil; current = current.Inside {
		if current.CarriedBy == nil {
			continue
		}

		if current.CarriedBy.Flags&CHAR_IS_PLAYER != 0 {
			return current.CarriedBy
		}

		return nil
	}

	return nil
}

func (game *Game) deletePersistedObjectInstance(obj *ObjectInstance) error {
	if obj == nil || obj.Id == 0 {
		return nil
	}

	err := game.deletePersistedObjectInstanceIDs([]uint{obj.Id})
	if err != nil {
		return err
	}

	obj.Id = 0
	return nil
}

func (game *Game) deletePersistedObjectTree(obj *ObjectInstance) error {
	if obj == nil {
		return nil
	}

	err := game.deletePersistedObjectInstanceIDs(obj.objectInstanceIDs())
	if err != nil {
		return err
	}

	obj.resetObjectInstanceIDs()
	return nil
}

func (game *Game) deletePersistedObjectInstanceIDs(ids []uint) error {
	if game == nil || game.db == nil {
		return nil
	}

	ids = compactObjectInstanceIDs(ids)
	if len(ids) == 0 {
		return nil
	}

	ctx := context.Background()
	tx, err := game.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = deleteObjectInstancesTx(ctx, tx, ids)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (obj *ObjectInstance) Finalize(container *ObjectInstance) error {
	if obj == nil || obj.Id > 0 {
		return nil
	}

	obj.ensureDecayState(time.Now())

	var insideObjectInstanceId *uint = nil

	if container != nil {
		insideObjectInstanceId = &container.Id
	}

	result, err := obj.Game.db.Exec(`
		INSERT INTO
			object_instances(parent_id, inside_object_instance_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4, weight, ttl, created_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, obj.ParentId, insideObjectInstanceId, obj.Name, obj.ShortDescription, obj.LongDescription, obj.Description, obj.Flags, obj.ItemType, obj.Value0, obj.Value1, obj.Value2, obj.Value3, obj.GetWeight(), obj.Ttl, obj.CreatedAt)
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

func (obj *ObjectInstance) ensureContents() *LinkedList {
	if obj.Contents == nil {
		obj.Contents = NewLinkedList()
	}

	return obj.Contents
}

func (container *ObjectInstance) AddObject(obj *ObjectInstance) {
	container.ensureContents().Insert(obj)

	obj.Inside = container
	obj.CarriedBy = nil
	obj.InRoom = nil
	obj.WearLocation = -1
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
