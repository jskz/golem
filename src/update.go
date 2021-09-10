/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

func (game *Game) Update() {
	for zone := range game.zones {
		game.ResetZone(zone)
	}
}
