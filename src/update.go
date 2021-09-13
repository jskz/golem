/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "time"

func (game *Game) Update() {
	for iter := game.characters.head; iter != nil; iter = iter.next {
		ch := iter.value.(*Character)

		ch.onUpdate()
	}
}

func (game *Game) ZoneUpdate() {
	for iter := game.zones.head; iter != nil; iter = iter.next {
		zone := iter.value.(*Zone)

		if time.Since(zone.lastReset).Minutes() > float64(zone.resetFrequency) {
			game.ResetZone(zone)
		}
	}
}
