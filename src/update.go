/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

func (game *Game) Update() {
}

func (game *Game) ZoneUpdate() {
	for iter := game.zones.head; iter != nil; iter = iter.next {
		zone := iter.value.(*Zone)

		game.ResetZone(zone)
	}
}
