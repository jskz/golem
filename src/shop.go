/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "log"

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
