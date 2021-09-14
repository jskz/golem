/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"time"
)

func (game *Game) Update() {
	game.InvokeNamedEventHandlersWithContextAndArguments("gameUpdate", game.vm.ToValue(game))

	for iter := game.characters.Head; iter != nil; iter = iter.Next {
		ch := iter.Value.(*Character)

		ch.onUpdate()
	}
}

func (game *Game) ZoneUpdate() {
	for iter := game.zones.Head; iter != nil; iter = iter.Next {
		zone := iter.Value.(*Zone)

		if time.Since(zone.lastReset).Minutes() > float64(zone.resetFrequency) {
			game.ResetZone(zone)
			game.InvokeNamedEventHandlersWithContextAndArguments("zoneUpdate", game.vm.ToValue(zone))
		}
	}
}
