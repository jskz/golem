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

func do_equipment(ch *Character, arguments string) {
	var output strings.Builder

	output.WriteString("\r\n{WYou are equipped with the following:{x\r\n")

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

	var found *ObjectInstance = nil
	for iter := ch.room.objects.head; iter != nil; iter = iter.next {
		obj := iter.value.(*ObjectInstance)

		/*
		 * TODO: add method on character to lookup prefix/name on inventory.
		 * Implement familiar syntax for indexing: take 2.sword, drop 3.potion, etc.
		 */
		if strings.Contains(obj.name, arguments) {
			found = obj
		}
	}

	if found == nil {
		ch.Send("No such item found.\r\n")
		return
	}

	/* TODO: Check if object can be taken, weight limits, etc */

	ch.room.removeObject(found)
	ch.addObject(found)

	ch.Send(fmt.Sprintf("You take %s{x.\r\n", found.shortDescription))
	outString := fmt.Sprintf("\r\n%s takes %s{x.\r\n", ch.name, found.shortDescription)

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

	var found *ObjectInstance = nil
	for iter := ch.inventory.head; iter != nil; iter = iter.next {
		obj := iter.value.(*ObjectInstance)

		/*
		 * TODO: add method on character to lookup prefix/name on inventory.
		 * Implement familiar syntax for indexing: take 2.sword, drop 3.potion, etc.
		 */
		if strings.Contains(obj.name, arguments) {
			found = obj
		}
	}

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