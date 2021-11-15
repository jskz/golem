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
	"math"
	"strconv"
	"strings"
	"time"
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
			var buf strings.Builder

			if obj.Flags&ITEM_GLOW != 0 {
				buf.WriteString("{G(Glowing){x ")
			}

			if obj.Flags&ITEM_HUM != 0 {
				buf.WriteString("{M(Humming){x ")
			}

			buf.WriteString(obj.GetShortDescription(ch))
			description = buf.String()
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

	output.WriteString(fmt.Sprintf("{cObject {C'%s'{c is type {C%s{c with flags {C%s{c.\r\n", obj.Name, obj.ItemType, obj.GetFlagsString()))
	output.WriteString(fmt.Sprintf("{C%s{x\r\n", obj.Description))

	if obj.Flags&ITEM_DECAYS != 0 {
		var silentString string = " "
		var minuteSuffix string = "s"
		var minutesSince int = int(time.Since(obj.CreatedAt).Minutes())
		var ttlSuffix string = "s"

		if obj.Flags&ITEM_DECAY_SILENTLY != 0 {
			silentString = " silently "
		}

		if minutesSince == 1 {
			minuteSuffix = ""
		}

		objExpiry := int(math.Max(time.Until(obj.CreatedAt.Add(time.Duration(obj.Ttl)*time.Minute)).Minutes(), 0))
		if objExpiry == 1 {
			ttlSuffix = ""
		}

		output.WriteString(fmt.Sprintf("{Y* {C%s{c was created %d minute%s ago and will%svanish in {C%d{c minute%s.\r\n",
			obj.GetShortDescriptionUpper(ch),
			int(time.Since(obj.CreatedAt).Minutes()),
			minuteSuffix,
			silentString,
			objExpiry,
			ttlSuffix,
		))
	}

	switch obj.ItemType {
	case ItemTypeArmor:
		output.WriteString(fmt.Sprintf("{Y* {C%s{c provides {C%d{c defense against bash damage.{x\r\n", obj.GetShortDescriptionUpper(ch), obj.Value0))
		output.WriteString(fmt.Sprintf("{Y* {C%s{c provides {C%d{c defense against slash damage.{x\r\n", obj.GetShortDescriptionUpper(ch), obj.Value1))
		output.WriteString(fmt.Sprintf("{Y* {C%s{c provides {C%d{c defense against piercing damage.{x\r\n", obj.GetShortDescriptionUpper(ch), obj.Value2))
		output.WriteString(fmt.Sprintf("{Y* {C%s{c provides {C%d{c defense against exotic damage.{x\r\n", obj.GetShortDescriptionUpper(ch), obj.Value3))
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

func (ch *Character) GetEquipment(wearLocation int) *ObjectInstance {
	for iter := ch.Inventory.Head; iter != nil; iter = iter.Next {
		obj := iter.Value.(*ObjectInstance)

		if obj.WearLocation == wearLocation {
			return obj
		}
	}

	return nil
}

func (ch *Character) detachEquipment(obj *ObjectInstance) bool {
	if ch.GetEquipment(obj.WearLocation) == nil {
		return false
	}

	obj.WearLocation = -1
	return true
}

func (ch *Character) attachEquipment(obj *ObjectInstance, wearLocation int) bool {
	if ch.GetEquipment(wearLocation) != nil {
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
		var obj *ObjectInstance = ch.GetEquipment(i)

		if obj == nil {
			objectDescription.WriteString("nothing")
		} else {
			objectDescription.WriteString(obj.GetShortDescription(ch))
		}

		output.WriteString(fmt.Sprintf("{C%s{x%s{x\r\n", WearLocations[i], objectDescription.String()))
	}

	ac := ch.GetArmorValues()

	output.WriteString(fmt.Sprintf("\r\n{WArmor versus bash damage:      %d\r\n", ac[DamageTypeBash]))
	output.WriteString(fmt.Sprintf("Armor versus slash damage:     %d\r\n", ac[DamageTypeSlash]))
	output.WriteString(fmt.Sprintf("Armor versus piercing damage:  %d\r\n", ac[DamageTypeStab]))
	output.WriteString(fmt.Sprintf("Armor versus exotic damage:    %d{x\r\n", ac[DamageTypeExotic]))

	ch.Send(output.String())
}

func do_inventory(ch *Character, arguments string) {
	var weightTotal float64 = 0.0

	ch.Send("\r\n{YYour current inventory:{x\r\n")
	ch.listObjects(ch.Inventory, false, true)

	if ch.Gold > 0 {
		var goldPlural string = ""
		if ch.Gold != 1 {
			goldPlural = "s"
		}

		ch.Send(fmt.Sprintf("{xYou are carrying {Y%d gold coin%s{x.\r\n", ch.Gold, goldPlural))
	}

	ch.Send(fmt.Sprintf("{xTotal: %d/%d items, %0.1f/%.1f lbs.\r\n",
		ch.Inventory.Count,
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
					wearing := ch.GetEquipment(WearLocationHeld)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop holding %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't let go of %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationWielded)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wielding %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wielding %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationBody)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationHead)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationNeck)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationTorso)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationArms)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationHands)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationWaist)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationShield)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationLegs)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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
					wearing := ch.GetEquipment(WearLocationFeet)
					if wearing != nil {
						result := ch.detachEquipment(wearing)
						if result {
							ch.Send(fmt.Sprintf("You stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
						} else {
							ch.Send(fmt.Sprintf("You can't stop wearing %s{x.\r\n", wearing.GetShortDescription(ch)))
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

		if ch.Inventory.Count+1 > ch.getMaxItemsInventory() {
			ch.Send("You can't carry any more.\r\n")
			return
		}

		if takingObj.ItemType != ItemTypeCurrency {
			err := ch.attachObject(takingObj)
			if err != nil {
				ch.Send(fmt.Sprintf("A strange force prevents you from removing %s from %s.\r\n", takingObj.GetShortDescription(ch), takingFrom.GetShortDescription(ch)))
				return
			}
		}

		takingFrom.removeObject(takingObj)

		if takingObj.ItemType != ItemTypeCurrency {
			ch.addObject(takingObj)
		} else {
			ch.Gold = ch.Gold + takingObj.Value0
			ch.Game.Objects.Remove(takingObj)
		}

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

	if ch.Inventory.Count+1 > ch.getMaxItemsInventory() {
		ch.Send("You can't carry any more.\r\n")
		return
	}

	/* TODO: Check if object can be taken, weight limits, etc */
	if ch.Flags&CHAR_IS_PLAYER != 0 {
		if found.ItemType != ItemTypeCurrency {
			err := ch.attachObject(found)
			if err != nil {
				log.Println(err)
				ch.Send("A strange force prevents you from taking that.\r\n")
				return
			}

		}

		ch.Room.removeObject(found)
	} else {
		ch.Room.removeObject(found)
	}

	if found.ItemType != ItemTypeCurrency {
		ch.addObject(found)
	} else {
		ch.Gold = ch.Gold + found.Value0
		ch.Game.Objects.Remove(found)
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

	firstArgument, arguments := oneArgument(arguments)
	secondArgument, arguments := oneArgument(arguments)
	thirdArgument, _ := oneArgument(arguments)

	if ch.Room == nil {
		return
	}

	if secondArgument == "gold" {
		amount, err := strconv.Atoi(firstArgument)
		if err != nil {
			ch.Send("Please provide an integer amount of gold to give and another player to give it to.\r\n")
			return
		}

		var target *Character = ch.FindCharacterInRoom(thirdArgument)
		if target == nil {
			ch.Send("No such person here.\r\n")
			return
		}

		if amount > ch.Gold {
			ch.Send("You don't have enough gold.\r\n")
			return
		}

		/* We don't need to exchange the object, but it is handy to have its short description */
		goldRepresentation := ch.Game.CreateGold(amount)

		ch.Gold -= amount
		target.Gold += amount

		ch.Send(fmt.Sprintf("You give %s{x to %s{x.\r\n", goldRepresentation.GetShortDescription(ch), target.GetShortDescription(ch)))
		target.Send(fmt.Sprintf("%s{x gives you %s{x.\r\n", ch.GetShortDescriptionUpper(target), goldRepresentation.GetShortDescription(target)))

		if ch.Room != nil {
			for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
				rch := iter.Value.(*Character)

				if rch != ch && rch != target {
					rch.Send(fmt.Sprintf("\r\n%s{x gives %s{x to %s{x.\r\n", ch.GetShortDescriptionUpper(rch), goldRepresentation.GetShortDescription(rch), target.GetShortDescription(rch)))
				}
			}
		}

		return
	}

	var found *ObjectInstance = ch.findObjectOnSelf(firstArgument)
	if found == nil {
		ch.Send("No such item in your inventory.\r\n")
		return
	}

	var target *Character = ch.FindCharacterInRoom(secondArgument)
	if target == nil {
		ch.Send("No such person here.\r\n")
		return
	}

	if target == ch {
		ch.Send("You cannot give to yourself!\r\n")
		return
	}

	if ch.Flags&CHAR_IS_PLAYER != 0 {
		err := ch.DetachObject(found)
		if err != nil {
			ch.Send("A strange force prevents you from releasing your grip.\r\n")
			return
		}

		ch.RemoveObject(found)
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

	firstArgument, arguments := oneArgument(arguments)
	secondArgument, _ := oneArgument(arguments)

	if secondArgument == "gold" {
		amount, err := strconv.Atoi(firstArgument)
		if err != nil {
			ch.Send("Please provide a valid integer gold amount.\r\n")
			return
		}

		if amount > ch.Gold {
			ch.Send("You don't have enough gold.\r\n")
			return
		}

		ch.Gold -= amount

		var found *ObjectInstance = nil

		for iter := ch.Room.Objects.Head; iter != nil; iter = iter.Next {
			obj := iter.Value.(*ObjectInstance)

			if obj.ItemType == ItemTypeCurrency {
				found = obj
				break
			}
		}

		gold := ch.Game.CreateGold(amount)
		ch.Send(fmt.Sprintf("You drop %s.\r\n", gold.GetShortDescription(ch)))

		if found != nil {
			amount += found.Value0

			ch.Room.Objects.Remove(found)
			ch.Game.Objects.Remove(found)

			gold = ch.Game.CreateGold(amount)
		}

		ch.Room.Objects.Insert(gold)
		ch.Game.Objects.Insert(gold)

		for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
			rch := iter.Value.(*Character)

			if !rch.IsEqual(ch) {
				rch.Send(fmt.Sprintf("\r\n%s drops %s{x.\r\n", ch.GetShortDescriptionUpper(rch), gold.GetShortDescription(rch)))
			}
		}

		return
	}

	var found *ObjectInstance = ch.findObjectOnSelf(firstArgument)
	if found == nil {
		ch.Send("No such item in your inventory.\r\n")
		return
	}

	if ch.Flags&CHAR_IS_PLAYER != 0 {
		err := ch.DetachObject(found)
		if err != nil {
			ch.Send("A strange force prevents you from releasing your grip.\r\n")
			return
		}

		ch.RemoveObject(found)
		ch.Room.addObject(found)
	} else {
		ch.RemoveObject(found)
		ch.Room.addObject(found)
	}

	ch.Send(fmt.Sprintf("You drop %s{x.\r\n", found.ShortDescription))
	outString := fmt.Sprintf("\r\n%s drops %s{x.\r\n", ch.Name, found.ShortDescription)

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch != ch {
			rch.Send(outString)
		}
	}
}
