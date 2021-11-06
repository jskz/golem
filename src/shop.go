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
	game.mobileShops = make(map[uint]*Shop)

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
		shop := &Shop{Game: game, Listings: NewLinkedList()}
		err := rows.Scan(&shop.Id, &shop.MobileId)
		if err != nil {
			log.Printf("Unable to scan shop: %v.\r\n", err)
			return err
		}

		game.shops[uint(shop.Id)] = shop
		game.mobileShops[shop.MobileId] = shop
	}

	log.Print("Loading shop-object relations.\r\n")
	rows, err = game.db.Query(`
		SELECT
			id,
			price,
			shop_id,
			object_id
		FROM
			shop_object
	`)
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	var objectIds map[uint]int = make(map[uint]int)

	for rows.Next() {
		var shopId uint
		var objectId uint

		shopListing := &ShopListing{}
		err := rows.Scan(&shopListing.Id, &shopListing.Price, &shopId, &objectId)
		if err != nil {
			log.Printf("Unable to scan shop: %v.\r\n", err)
			return err
		}

		_, ok := game.shops[shopId]
		if !ok {
			continue
		}

		shopListing.Shop = game.shops[shopId]
		shopListing.Shop.Listings.Insert(shopListing.Shop)
		objectIds[objectId] = shopListing.Id
	}

	/*
	 * At this point, any shop listings have been loaded but have not had their
	 * object structure instances hydrated.  We will try to bulk load every ID
	 * and then populate any shop listing with its
	 */
	var ids []uint = make([]uint, 0)

	for id := range objectIds {
		ids = append(ids, id)
	}

	objects, err := game.LoadObjectsByIndices(ids)
	if err != nil {
		return err
	}

	var objectFromId map[uint]*Object = make(map[uint]*Object)
	for _, obj := range objects {
		objectFromId[obj.Id] = obj
	}

	for _, shop := range game.shops {
		for iter := shop.Listings.Head; iter != nil; iter = iter.Next {
			listing := iter.Value.(*ShopListing)

			objectId, ok := objectIds[uint(listing.Id)]
			if !ok {
				continue
			}

			listing.Object, ok = objectFromId[uint(objectId)]
			if !ok {
				shop.Listings.Remove(listing)
				continue
			}
		}
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
			shop, ok := ch.Game.mobileShops[uint(rch.Id)]
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
	shop := ch.FindShopInRoom()
	if shop == nil {
		ch.Send("You can't do that here.\r\n")
		return
	}
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
