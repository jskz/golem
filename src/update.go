/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"fmt"
	"log"
	"time"
)

func (game *Game) characterUpdate() {
	for iter := game.Characters.Head; iter != nil; iter = iter.Next {
		ch := iter.Value.(*Character)

		if ch.Casting != nil {
			ch.onCastingUpdate()
		}

		for effectIter := ch.Effects.Head; effectIter != nil; effectIter = effectIter.Next {
			fx := effectIter.Value.(*Effect)

			if int(time.Since(fx.CreatedAt).Seconds()) >= fx.Duration {
				if fx.OnComplete != nil {
					_, err := (*fx.OnComplete)(game.vm.ToValue(ch))
					if err != nil {
						log.Println(err)
					}
				}

				ch.RemoveEffect(fx)
			}
		}
	}
}

func (game *Game) objectUpdate() {
	for iter := game.Objects.Head; iter != nil; {
		next := iter.Next
		obj := iter.Value.(*ObjectInstance)

		/* Remove the obj after its ttl time in minutes, if the ITEM_DECAYS flag is set */
		if obj.Flags&ITEM_DECAYS != 0 && int(time.Since(obj.CreatedAt).Minutes()) >= obj.Ttl {
			game.decayObject(obj)
		}

		iter = next
	}
}

type objectLocation struct {
	room      *Room
	carrier   *Character
	container *ObjectInstance
}

func (obj *ObjectInstance) location() objectLocation {
	return objectLocation{
		room:      obj.InRoom,
		carrier:   obj.CarriedBy,
		container: obj.Inside,
	}
}

func (game *Game) decayObject(obj *ObjectInstance) {
	location := obj.location()

	game.sendDecayMessage(obj, location)

	if obj.ItemType == ItemTypeContainer && obj.Contents != nil {
		game.moveDecayedContainerContents(obj, location)
	}

	obj.removeFromLocation()
	game.Objects.Remove(obj)
}

func (game *Game) sendDecayMessage(obj *ObjectInstance, location objectLocation) {
	if obj.Flags&ITEM_DECAY_SILENTLY != 0 {
		return
	}

	if location.room != nil {
		for innerIter := location.room.Characters.Head; innerIter != nil; innerIter = innerIter.Next {
			rch := innerIter.Value.(*Character)

			rch.Send(fmt.Sprintf("{D%s{D crumbles into dust.{x\r\n", obj.GetShortDescriptionUpper(rch)))
		}

		return
	}

	if location.carrier != nil {
		location.carrier.Send(fmt.Sprintf("{D%s{D crumbles into dust.{x\r\n", obj.GetShortDescriptionUpper(location.carrier)))
	}
}

func (game *Game) moveDecayedContainerContents(container *ObjectInstance, location objectLocation) {
	for contentIter := container.Contents.Head; contentIter != nil; {
		next := contentIter.Next
		contentObj := contentIter.Value.(*ObjectInstance)

		container.removeObject(contentObj)
		game.moveDecayedContainerContent(contentObj, location)

		contentIter = next
	}
}

func (game *Game) moveDecayedContainerContent(obj *ObjectInstance, location objectLocation) {
	switch {
	case location.room != nil:
		game.moveDecayedContainerContentToRoom(obj, location.room)
	case location.carrier != nil:
		game.moveDecayedContainerContentToCarrier(obj, location.carrier)
	case location.container != nil:
		if location.container.Contents == nil {
			location.container.Contents = NewLinkedList()
		}

		location.container.AddObject(obj)
	default:
		game.Objects.Remove(obj)
	}
}

func (game *Game) moveDecayedContainerContentToRoom(obj *ObjectInstance, room *Room) {
	var found *ObjectInstance = nil

	for iter := room.Objects.Head; iter != nil; iter = iter.Next {
		roomObj := iter.Value.(*ObjectInstance)

		if roomObj.ItemType == ItemTypeCurrency {
			found = roomObj
			break
		}
	}

	if found != nil && obj.ItemType == ItemTypeCurrency {
		room.Objects.Remove(found)
		game.Objects.Remove(found)
		game.Objects.Remove(obj)

		obj = game.CreateGold(found.Value0 + obj.Value0)
		game.Objects.Insert(obj)
	}

	room.AddObject(obj)
}

func (game *Game) moveDecayedContainerContentToCarrier(obj *ObjectInstance, carrier *Character) {
	if obj.ItemType == ItemTypeCurrency {
		carrier.Gold += obj.Value0
		game.Objects.Remove(obj)
		return
	}

	carrier.AddObject(obj)
}

func (obj *ObjectInstance) removeFromLocation() {
	switch {
	case obj.InRoom != nil:
		obj.InRoom.removeObject(obj)
	case obj.CarriedBy != nil:
		obj.CarriedBy.RemoveObject(obj)
	case obj.Inside != nil:
		obj.Inside.removeObject(obj)
	default:
		obj.InRoom = nil
		obj.CarriedBy = nil
		obj.Inside = nil
	}
}

func (game *Game) Update() {
	for iter := game.Characters.Head; iter != nil; iter = iter.Next {
		ch := iter.Value.(*Character)

		ch.onUpdate()
	}
}

func (game *Game) ZoneUpdate() {
	for iter := game.Zones.Head; iter != nil; iter = iter.Next {
		zone := iter.Value.(*Zone)

		if time.Since(zone.LastReset).Minutes() > float64(zone.ResetFrequency) {
			game.ResetZone(zone)
		}
	}

	/* Run district update scripts */
	for districtId, script := range game.districtScripts {
		district := game.FindDistrictByID(districtId)
		if district == nil {
			log.Printf("Couldn't run district-script for nonexistent district id %d.\r\n", districtId)
			continue
		}

		_, err := script.tryEvaluate("onUpdate", game.vm.ToValue(district))
		if err != nil {
			log.Printf("Script evaluation of %d for district %d onUpdate failed: %v\r\n", script.Id, districtId, err)
		}
	}
}
