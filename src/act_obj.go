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
	"strings"
)

const (
	WearLocationNone    = 0
	WearLocationHead    = 1
	WearLocationNeck    = 2
	WearLocationArms    = 3
	WearLocationTorso   = 4
	WearLocationLegs    = 5
	WearLocationHands   = 6
	WearLocationShield  = 7
	WearLocationBody    = 8
	WearLocationWaist   = 9
	WearLocationWielded = 10
	WearLocationHeld    = 11
	WearLocationMax     = 12
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
	WearLocations[WearLocationWielded] = "<wielded>             "
	WearLocations[WearLocationHeld] = "<held>                "
}

func do_equipment(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString("\r\n{WYou are equipped with the following:{x\r\n")

	for i := WearLocationNone + 1; i < WearLocationMax; i++ {
		var objectDescription strings.Builder

		if ch.equipment[i] == nil {
			objectDescription.WriteString("nothing")
		} else {
			obj := ch.equipment[i]

			objectDescription.WriteString(obj.shortDescription)

			/* TODO: item flags - glowing, humming, etc? Append extra details here. */
		}

		output.WriteString(fmt.Sprintf("{C%s{x%s{x\r\n", WearLocations[i], objectDescription.String()))
	}

	ch.Send(output.String())
}

func do_inventory(ch *Character, arguments string) {
	var output strings.Builder
	var count int = 0
	var weightTotal float64 = 0.0

	output.WriteString("\r\n{YYour current inventory:{x\r\n")

	for iter := ch.inventory.head; iter != nil; iter = iter.next {
		obj := iter.value.(*ObjectInstance)

		output.WriteString(fmt.Sprintf("{x    %s\r\n", obj.shortDescription))

		count++
	}

	output.WriteString(fmt.Sprintf("{xTotal: %d/%d items, %0.1f/%.1f lbs.\r\n",
		count,
		ch.getMaxItemsInventory(),
		weightTotal,
		ch.getMaxCarryWeight()))

	ch.Send(output.String())
}

func do_wear(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Wear what?\r\n")
		return
	}

	ch.Send("Not yet implemented, try again soon!\r\n")
}

func do_remove(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Remove what?\r\n")
		return
	}

	ch.Send("Not yet implemented, try again soon!\r\n")
}

func do_take(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Take what?\r\n")
		return
	}

	if ch.room == nil {
		return
	}

	var found *ObjectInstance = ch.findObjectInRoom(arguments)
	if found == nil {
		ch.Send("No such item found.\r\n")
		return
	}

	/* TODO: Check if object can be taken, weight limits, etc */

	ch.room.removeObject(found)
	ch.addObject(found)

	ch.Send(fmt.Sprintf("You take %s{x.\r\n", found.shortDescription))
	outString := fmt.Sprintf("\r\n%s{x takes %s{x.\r\n", ch.name, found.shortDescription)

	if ch.room != nil {
		for iter := ch.room.characters.head; iter != nil; iter = iter.next {
			rch := iter.value.(*Character)

			if rch != ch {
				rch.Send(outString)
			}
		}
	}
}

func do_drop(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Drop what?\r\n")
		return
	}

	if ch.room == nil {
		return
	}

	var found *ObjectInstance = ch.findObjectOnSelf(arguments)
	if found == nil {
		ch.Send("No such item in your inventory.\r\n")
		return
	}

	ch.removeObject(found)
	ch.room.addObject(found)

	ch.Send(fmt.Sprintf("You drop %s{x.\r\n", found.shortDescription))
	outString := fmt.Sprintf("\r\n%s drops %s{x.\r\n", ch.name, found.shortDescription)

	if ch.room != nil {
		for iter := ch.room.characters.head; iter != nil; iter = iter.next {
			rch := iter.value.(*Character)

			if rch != ch {
				rch.Send(outString)
			}
		}
	}
}
