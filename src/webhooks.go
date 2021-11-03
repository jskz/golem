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
	"net/http"
)

type Webhook struct {
	Id   int    `json:"id"`
	Uuid string `json:"uuid"`
}

func (game *Game) LoadWebhooks() error {
	log.Printf("Loading webhooks.\r\n")

	game.webhooks = make(map[string]*Webhook)

	rows, err := game.db.Query(`
		SELECT
			id,
			uuid
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

		err := rows.Scan(&webhook.Id, &webhook.Uuid)
		if err != nil {
			log.Printf("Unable to scan webhook: %v.\r\n", err)
			return err
		}

		game.webhooks[webhook.Uuid] = webhook
	}

	return nil
}

func (game *Game) handleWebhooks() {
	http.HandleFunc("/webhook", func(w http.ResponseWriter, req *http.Request) {
		log.Printf("Got a webhook request with key: %s\r\n", req.URL.Query().Get("key"))
	})

	http.ListenAndServe(":9000", nil)
}
