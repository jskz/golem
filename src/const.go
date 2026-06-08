/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

const LiquidWater = 0

type Liquid struct {
	Name   string
	Colour string

	Drunk  int
	Full   int
	Thirst int
	Hunger int
	Sip    int
}

var LiquidTable = []Liquid{
	{Name: "water", Colour: "clear", Drunk: 0, Full: 1, Thirst: 10, Hunger: 0, Sip: 16},
	{Name: "beer", Colour: "amber", Drunk: 12, Full: 1, Thirst: 8, Hunger: 1, Sip: 12},
	{Name: "red wine", Colour: "burgundy", Drunk: 30, Full: 1, Thirst: 8, Hunger: 1, Sip: 5},
	{Name: "ale", Colour: "brown", Drunk: 15, Full: 1, Thirst: 8, Hunger: 1, Sip: 12},
	{Name: "dark ale", Colour: "dark", Drunk: 16, Full: 1, Thirst: 8, Hunger: 1, Sip: 12},
	{Name: "whisky", Colour: "golden", Drunk: 120, Full: 1, Thirst: 5, Hunger: 0, Sip: 2},
	{Name: "lemonade", Colour: "pink", Drunk: 0, Full: 1, Thirst: 9, Hunger: 2, Sip: 12},
	{Name: "firebreather", Colour: "boiling", Drunk: 190, Full: 0, Thirst: 4, Hunger: 0, Sip: 2},
	{Name: "local specialty", Colour: "clear", Drunk: 151, Full: 1, Thirst: 3, Hunger: 0, Sip: 2},
	{Name: "slime mold juice", Colour: "green", Drunk: 0, Full: 2, Thirst: -8, Hunger: 1, Sip: 2},
	{Name: "milk", Colour: "white", Drunk: 0, Full: 2, Thirst: 9, Hunger: 3, Sip: 12},
	{Name: "tea", Colour: "tan", Drunk: 0, Full: 1, Thirst: 8, Hunger: 0, Sip: 6},
	{Name: "coffee", Colour: "black", Drunk: 0, Full: 1, Thirst: 8, Hunger: 0, Sip: 6},
	{Name: "blood", Colour: "red", Drunk: 0, Full: 2, Thirst: -1, Hunger: 2, Sip: 6},
	{Name: "salt water", Colour: "clear", Drunk: 0, Full: 1, Thirst: -2, Hunger: 0, Sip: 1},
	{Name: "coke", Colour: "brown", Drunk: 0, Full: 2, Thirst: 9, Hunger: 2, Sip: 12},
	{Name: "root beer", Colour: "brown", Drunk: 0, Full: 2, Thirst: 9, Hunger: 2, Sip: 12},
	{Name: "elvish wine", Colour: "green", Drunk: 35, Full: 2, Thirst: 8, Hunger: 1, Sip: 5},
	{Name: "white wine", Colour: "golden", Drunk: 28, Full: 1, Thirst: 8, Hunger: 1, Sip: 5},
	{Name: "champagne", Colour: "golden", Drunk: 32, Full: 1, Thirst: 8, Hunger: 1, Sip: 5},
	{Name: "mead", Colour: "honey-colored", Drunk: 34, Full: 2, Thirst: 8, Hunger: 2, Sip: 12},
	{Name: "rose wine", Colour: "pink", Drunk: 26, Full: 1, Thirst: 8, Hunger: 1, Sip: 5},
	{Name: "benedictine wine", Colour: "burgundy", Drunk: 40, Full: 1, Thirst: 8, Hunger: 1, Sip: 5},
	{Name: "vodka", Colour: "clear", Drunk: 130, Full: 1, Thirst: 5, Hunger: 0, Sip: 2},
	{Name: "cranberry juice", Colour: "red", Drunk: 0, Full: 1, Thirst: 9, Hunger: 2, Sip: 12},
	{Name: "orange juice", Colour: "orange", Drunk: 0, Full: 2, Thirst: 9, Hunger: 3, Sip: 12},
	{Name: "absinthe", Colour: "green", Drunk: 200, Full: 1, Thirst: 4, Hunger: 0, Sip: 2},
	{Name: "brandy", Colour: "golden", Drunk: 80, Full: 1, Thirst: 5, Hunger: 0, Sip: 4},
	{Name: "aquavit", Colour: "clear", Drunk: 140, Full: 1, Thirst: 5, Hunger: 0, Sip: 2},
	{Name: "schnapps", Colour: "clear", Drunk: 90, Full: 1, Thirst: 5, Hunger: 0, Sip: 2},
	{Name: "icewine", Colour: "purple", Drunk: 50, Full: 2, Thirst: 6, Hunger: 1, Sip: 5},
	{Name: "amontillado", Colour: "burgundy", Drunk: 35, Full: 2, Thirst: 8, Hunger: 1, Sip: 5},
	{Name: "sherry", Colour: "red", Drunk: 38, Full: 2, Thirst: 7, Hunger: 1, Sip: 5},
	{Name: "framboise", Colour: "red", Drunk: 50, Full: 1, Thirst: 7, Hunger: 1, Sip: 5},
	{Name: "rum", Colour: "amber", Drunk: 151, Full: 1, Thirst: 4, Hunger: 0, Sip: 2},
	{Name: "cordial", Colour: "clear", Drunk: 100, Full: 1, Thirst: 5, Hunger: 0, Sip: 2},
}

func normalizeLiquid(liquid int) int {
	if liquid < 0 || liquid >= len(LiquidTable) {
		return LiquidWater
	}

	return liquid
}
