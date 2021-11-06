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
	"strings"
)

type ShopListing struct {
	Shop   *Shop   `json:"shop"`
	Id     int     `json:"id"`
	Object *Object `json:"object"`
	Price  int     `json:"price"`
}

type Shop struct {
	Game     *Game       `json:"game"`
	Id       int         `json:"id"`
	MobileId uint        `json:"mobileId"`
	Listings *LinkedList `json:"listings"`
}

func (game *Game) LoadShops() error {
	log.Printf("Loading shops.\r\n")

	game.shops = make(map[uint]*Shop)

	rows, err := game.db.Query(`
		SELECT
			id,
			mobile_id
		FROM
			shops
	`)
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		shop := &Shop{Game: game}

		err := rows.Scan(&shop.Id, &shop.MobileId)
		if err != nil {
			log.Printf("Unable to scan shop: %v.\r\n", err)
			return err
		}

		game.shops[shop.MobileId] = shop
	}

	return nil
}

func (shop *Shop) Save() error {
	return nil
}

func (ch *Character) FindShopInRoom() *Shop {
	if ch == nil || ch.Room == nil || ch.Room.Characters == nil {
		return nil
	}

	for iter := ch.Room.Characters.Head; iter != nil; iter = iter.Next {
		rch := iter.Value.(*Character)

		if rch.Flags&CHAR_SHOPKEEPER != 0 {
			shop, ok := ch.Game.shops[uint(rch.Id)]
			if !ok {
				/* Flagged shopkeeper but no associated shop */
				continue
			}

			return shop
		}
	}

	return nil
}

func do_buy(ch *Character, arguments string) {

}

func do_shop(ch *Character, arguments string) {
	var output strings.Builder
	var count int = 1

	shop := ch.FindShopInRoom()
	if shop == nil {
		ch.Send("You can't do that here.\r\n")
		return
	}

	for iter := shop.Listings.Head; iter != nil; iter = iter.Next {
		listing := iter.Value.(*ShopListing)

		output.WriteString(fmt.Sprintf("%2d) %-32s %5d gold coins{x\r\n", count, listing.Object.ShortDescription, listing.Price))
		count++
	}

	ch.Send(output.String())
}
