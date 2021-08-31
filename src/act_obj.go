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
}

func do_drop(ch *Character, arguments string) {
	if len(arguments) < 1 {
		ch.Send("Drop what?\r\n")
		return
	}

	ch.Send("Not yet implemented, try again soon!\r\n")
}
