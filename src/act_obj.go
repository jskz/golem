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
	"strings"
)

const (
	WearLocationNone    = 0
	WearLocationHead    = 1
	WearLocationNeck    = 2
	WearLocationArms    = 3
	WearLocationTorso   = 4
	WearLocationHands   = 5
	WearLocationShield  = 6
	WearLocationBody    = 7
	WearLocationWaist   = 8
	WearLocationLegs    = 9
	WearLocationFeet    = 10
	WearLocationWielded = 11
	WearLocationHeld    = 12
	WearLocationMax     = 13
)

var WearLocations = make(map[int]string)

func init() {
	/* Initialize our wear location string map */
	WearLocations[WearLocationNone] = ""
	WearLocations[WearLocationHead] = "<worn on head>        "
	WearLocations[WearLocationNeck] = "<worn around neck>    "
	WearLocations[WearLocationArms] = "<worn on arms>        "
	WearLocations[WearLocationTorso] = "<worn on torso>       "
	WearLocations[WearLocationLegs] = "<worn on legs>        "
	WearLocations[WearLocationHands] = "<worn on hands>       "
	WearLocations[WearLocationShield] = "<worn as shield>      "
	WearLocations[WearLocationBody] = "<worn on body>        "
	WearLocations[WearLocationWaist] = "<worn around waist>   "
	WearLocations[WearLocationFeet] = "<worn on feet>        "
	WearLocations[WearLocationWielded] = "<wielded>             "
	WearLocations[WearLocationHeld] = "<held>                "
}

func (ch *Character) listObjects(objects *LinkedList, longDescriptions bool, hideEquipped bool) {
	var output strings.Builder
	var inventory map[string]uint = make(map[string]uint)

	for iter := objects.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		if hideEquipped && obj.WearLocation != -1 {
			continue
		}

		var description string = obj.LongDescription
		if !longDescriptions {
			description = obj.ShortDescription
		}

		_, ok := inventory[description]
		if !ok {
			inventory[description] = 1
		} else {
			inventory[description]++
		}
	}

	for iter := objects.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		if hideEquipped && obj.WearLocation != -1 {
			continue
		}

		var description string = obj.LongDescription
		if !longDescriptions {
			description = obj.ShortDescription
		}

		count, ok := inventory[description]
		if !ok {
			continue
		}

		if count > 1 {
			output.WriteString(fmt.Sprintf("(%3d) %s{x\r\n", count, description))
			delete(inventory, description)
			continue
		}

		output.WriteString(fmt.Sprintf("      %s{x\r\n", description))
		delete(inventory, description)
	}

	ch.Send(output.String())
}

func (ch *Character) examineObject(obj *ObjectInstance) {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("{cObject {C'%s'{c is type {C%s{c.\r\n", obj.Name, obj.ItemType))
	output.WriteString(fmt.Sprintf("{C%s{x\r\n", obj.Description))

	switch obj.ItemType {
	case ItemTypeContainer:
		output.WriteString(fmt.Sprintf("{C%s{c can hold up to {C%d{c items and {C%d{c lbs.{x\r\n", obj.GetShortDescriptionUpper(ch), obj.Value0, obj.Value1))
	default:
		break
	}

	if obj.Contents != nil && obj.Contents.Count > 0 {
		output.WriteString(fmt.Sprintf("{C%s{c contains the following items:\r\n", obj.GetShortDescriptionUpper(ch)))
		ch.Send(output.String())

		ch.showObjectList(obj.Contents)
		return
	}

	ch.Send(output.String())
}

func (ch *Character) getEquipment(wearLocation int) *ObjectInstance {
	for iter := ch.Inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		if obj.WearLocation == wearLocation {
			return obj
		}
	}

	return nil
}

func (ch *Character) detachEquipment(obj *ObjectInstance) bool {
	if ch.getEquipment(obj.WearLocation) == nil {
		return false
	}

	obj.WearLocation = -1
	return true
}

func (ch *Character) attachEquipment(obj *ObjectInstance, wearLocation int) bool {
	if ch.getEquipment(wearLocation) != nil {
		return false
	}

	obj.WearLocation = wearLocation
	return true
}

func do_equipment(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString("\r\n{WYou are equipped with the following:{x\r\n")

	for i := WearLocationNone + 1; i < WearLocationMax; i++ {
		var objectDescription strings.Builder
		var obj *ObjectInstance = ch.getEquipment(i)

		if obj == nil {
			objectDescription.WriteString("nothing")
		} else {
			objectDescription.WriteString(obj.GetShortDescription(ch))
		}

		output.WriteString(fmt.Sprintf("{C%s{x%s{x\r\n", WearLocations[i], objectDescription.String()))
	}

	ch.Send(output.String())
}

func do_inventory(ch *Character, arguments string) {
	var count int = 0
	var weightTotal float64 = 0.0

	ch.Send("\r\n{YYour current inventory:{x\r\n")
	ch.listObjects(ch.Inventory, false, true)

	ch.Send(fmt.Sprintf("{xTotal: %d/%d items, %0.1f/%.1f lbs.\r\n",
		count,
		ch.getMaxItemsInventory(),
		weightTotal,
		ch.getMaxCarryWeight()))
}

func do_wear(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Wear what?\r\n")
		return
	}

	firstArgument, _ := oneArgument(arguments)

	for iter := ch.Inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		if obj.WearLocation == -1 {
			if strings.Contains(obj.Name, firstArgument) {
				if obj.Flags&ITEM_WEAR_HELD != 0 {
					wearing := ch.getEquipment(WearLocationHeld)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop holding %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't let go of %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationHeld) {
						ch.Send(fmt.Sprintf("You hold %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x holds %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAPON != 0 {
					wearing := ch.getEquipment(WearLocationWielded)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wielding %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wielding %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationWielded) {
						ch.Send(fmt.Sprintf("You wield %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wields %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_BODY != 0 {
					wearing := ch.getEquipment(WearLocationBody)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationBody) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_HEAD != 0 {
					wearing := ch.getEquipment(WearLocationHead)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationHead) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_NECK != 0 {
					wearing := ch.getEquipment(WearLocationNeck)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationNeck) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_TORSO != 0 {
					wearing := ch.getEquipment(WearLocationTorso)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationTorso) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_ARMS != 0 {
					wearing := ch.getEquipment(WearLocationArms)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationArms) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_HANDS != 0 {
					wearing := ch.getEquipment(WearLocationHands)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationHands) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_WAIST != 0 {
					wearing := ch.getEquipment(WearLocationWaist)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationWaist) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_SHIELD != 0 {
					wearing := ch.getEquipment(WearLocationShield)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationShield) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_LEGS != 0 {
					wearing := ch.getEquipment(WearLocationLegs)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationLegs) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				} else if obj.Flags&ITEM_WEAR_FEET != 0 {
					wearing := ch.getEquipment(WearLocationFeet)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", obj.GetShortDescription(ch)))
							return
						}
					}

					if ch.attachEquipment(obj, WearLocationFeet) {
						ch.Send(fmt.Sprintf("You wear %s{x.\r\n", obj.GetShortDescription(ch)))

						for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
							rch := roomOthersIter.Value.(*Character)

							if !rch.IsEqual(ch) {
								rch.Send(fmt.Sprintf("%s{x wears %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
							}
						}

						return
					}
				}
			}
		}
	}

	ch.Send("You can't wear that.\r\n")
}

func do_remove(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Remove what?\r\n")
		return
	}

	firstArgument, _ := oneArgument(arguments)

	for iter := ch.Inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		if obj.WearLocation != -1 {
			if strings.Contains(obj.Name, firstArgument) {
				result := ch.detachEquipment(obj)
				if result {
					ch.Send(fmt.Sprintf("You remove %s{x.\r\n", obj.GetShortDescription(ch)))

					for roomOthersIter := ch.Room.Characters.Head; roomOthersIter != nil; roomOthersIter = roomOthersIter.Next {
						rch := roomOthersIter.Value.(*Character)

						if !rch.IsEqual(ch) {
							rch.Send(fmt.Sprintf("%s{x removes %s{x.\r\n", ch.GetShortDescriptionUpper(rch), obj.GetShortDescription(rch)))
						}
					}

					return
				}

				ch.Send("A strange force prevents you from removing that.\r\n")
				return
			}
		}
	}

	ch.Send("You aren't wearing that.\r\n")
	return
}

func do_use(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Use what?\r\n")
		return
	}

	if ch.Room == nil {
		return
	}

	firstArgument, _ := oneArgument(arguments)
	var using *ObjectInstance = ch.findObjectInRoom(firstArgument)
	if using == nil {
		using = ch.findObjectOnSelf(firstArgument)

		if using == nil {
			ch.Send("No such item found.\r\n")
			return
		}
	}

	script, ok := using.Game.objectScripts[using.ParentId]
	if !ok {
		ch.Send("You can't use that.\r\n")
		return
	}

	_, err := script.tryEvaluate("onUse", ch.Game.vm.ToValue(using), ch.Game.vm.ToValue(ch))
	if err != nil {
		ch.Send("You can't use that.\r\n")
		return
	}
}

func do_take(ch *Character, arguments string) {
	var firstArgument string = ""
	var secondArgument string = ""

	if len(arguments) < 1 {
		ch.Send("Take what?\r\n")
		return
	}

	if ch.Room == nil {
		return
	}

	firstArgument, arguments = oneArgument(arguments)
	secondArgument, _ = oneArgument(arguments)

	if secondArgument != "" {
		/* Trying to take the object "firstArgument" from within the object "secondArgument" */
		var takingFrom *ObjectInstance = ch.findObjectInRoom(secondArgument)
		if takingFrom == nil {
			takingFrom = ch.findObjectOnSelf(secondArgument)
			if takingFrom == nil {
				ch.Send("No such item found.\r\n")
				return
			}
		}

		var takingObj *ObjectInstance = takingFrom.findObjectInSelf(ch, firstArgument)
		if takingObj == nil {
			ch.Send(fmt.Sprintf("No such item found in %s.\r\n", takingFrom.GetShortDescription(ch)))
			return
		}

		if takingObj.Flags&ITEM_TAKE == 0 {
			ch.Send(fmt.Sprintf("You are unable to take %s from %s.\r\n", takingObj.GetShortDescription(ch), takingFrom.GetShortDescription(ch)))
			return
		}

		err := ch.attachObject(takingObj)
		if err != nil {
			ch.Send(fmt.Sprintf("A strange force prevents you from removing %s from %s.\r\n", takingObj.GetShortDescription(ch), takingFrom.GetShortDescription(ch)))
			return
		}

		takingFrom.removeObject(takingObj)
		ch.addObject(takingObj)

		ch.Send(fmt.Sprintf("You take %s{x from %s{x.\r\n", takingObj.GetShortDescription(ch), takingFrom.GetShortDescription(ch)))
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch != ch {
				rch.Send(fmt.Sprintf("%s{x takes %s{x from %s{x.\r\n", ch.GetShortDescriptionUpper(rch), takingObj.GetShortDescription(rch), takingFrom.GetShortDescription(rch)))
			}
		}

		return
	}

	var found *ObjectInstance = ch.findObjectInRoom(firstArgument)
	if found == nil {
		ch.Send("No such item found.\r\n")
		return
	}

	if found.Flags&ITEM_TAKE == 0 {
		ch.Send("You can't take that.\r\n")
		return
	}

	/* TODO: Check if object can be taken, weight limits, etc */
	if ch.Flags&CHAR_IS_PLAYER != 0 {
		err := ch.attachObject(found)
		if err != nil {
			log.Println(err)
			ch.Send("A strange force prevents you from taking that.\r\n")
			return
		}

		ch.addObject(found)
		ch.Room.removeObject(found)
	} else {
		ch.addObject(found)
		ch.Room.removeObject(found)
	}

	ch.Send(fmt.Sprintf("You take %s{x.\r\n", found.ShortDescription))
	outString := fmt.Sprintf("\r\n%s{x takes %s{x.\r\n", ch.Name, found.ShortDescription)

	if ch.Room != nil {
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch != ch {
				rch.Send(outString)
			}
		}
	}
}

func do_give(ch *Character, arguments string) {
	args := strings.Split(arguments, " ")
	if len(args) < 2 {
		ch.Send("Give what to whom?\r\n")
		return
	}

	if ch.Room == nil {
		return
	}

	var found *ObjectInstance = ch.findObjectOnSelf(args[0])
	if found == nil {
		ch.Send("No such item in your inventory.\r\n")
		return
	}

	var target *Character = ch.FindCharacterInRoom(args[1])
	if target == nil {
		ch.Send("No such person here.\r\n")
		return
	}

	if target == ch {
		ch.Send("You cannot give to yourself!\r\n")
		return
	}

	if ch.Flags&CHAR_IS_PLAYER != 0 {
		err := ch.detachObject(found)
		if err != nil {
			ch.Send("A strange force prevents you from releasing your grip.\r\n")
			return
		}

		ch.removeObject(found)
	}

	if target.Flags&CHAR_IS_PLAYER != 0 {
		err := target.attachObject(found)
		if err != nil {
			ch.Send("A strange force prevents you from releasing your grip.\r\n")
			return
		}
	}

	target.addObject(found)

	ch.Send(fmt.Sprintf("You give %s{x to %s{x.\r\n", found.GetShortDescription(ch), target.GetShortDescription(ch)))
	target.Send(fmt.Sprintf("%s{x gives you %s{x.\r\n", ch.GetShortDescriptionUpper(target), found.GetShortDescription(target)))

	if ch.Room != nil {
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch != ch && rch != target {
				rch.Send(fmt.Sprintf("\r\n%s{x gives %s{x to %s{x.\r\n", ch.GetShortDescriptionUpper(rch), found.GetShortDescription(rch), target.GetShortDescription(rch)))
			}
		}
	}
}

func do_drop(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Drop what?\r\n")
		return
	}

	if ch.Room == nil {
		return
	}

	var found *ObjectInstance = ch.findObjectOnSelf(arguments)
	if found == nil {
		ch.Send("No such item in your inventory.\r\n")
		return
	}

	if ch.Flags&CHAR_IS_PLAYER != 0 {
		err := ch.detachObject(found)
		if err != nil {
			ch.Send("A strange force prevents you from releasing your grip.\r\n")
			return
		}

		ch.removeObject(found)
		ch.Room.addObject(found)
	} else {
		ch.removeObject(found)
		ch.Room.addObject(found)
	}

	ch.Send(fmt.Sprintf("You drop %s{x.\r\n", found.ShortDescription))
	outString := fmt.Sprintf("\r\n%s drops %s{x.\r\n", ch.Name, found.ShortDescription)

	if ch.Room != nil {
		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if rch != ch {
				rch.Send(outString)
			}
		}
	}
}
