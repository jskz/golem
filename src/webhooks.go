/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"log"
)

type Webhook struct {
	Id  int    `json:"id"`
	Url string `json:"url"`
}

func (game *Game) LoadWebhooks() error {
	log.Printf("Loading webhooks.\r\n")

	rows, err := game.db.Query(`
		SELECT
			id,
			url
		FROM
			webhooks
	`)
	if err != nil {
		log.Println(err)
		return err
	}

	defer rows.Close()

	for rows.Next() {
		webhook := &Webhook{}

		err := rows.Scan(&webhook.Id, &webhook.Url)
		if err != nil {
			log.Printf("Unable to scan webhook: %v.\r\n", err)
			return err
		}
	}

	return nil
}
