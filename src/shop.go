/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

type ShopListing struct {
	Shop   *Shop   `json:"shop"`
	Id     int     `json:"id"`
	Object *Object `json:"object"`
	Price  int     `json:"price"`
}

type Shop struct {
	Game     *Game       `json:"game"`
	Id       int         `json:"id"`
	MobileId int         `json:"mobileId"`
	Listings *LinkedList `json:"listings"`
}

func (game *Game) LoadShops() error {
	return nil
}

func (shop *Shop) Save() error {
	return nil
}
