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

	"github.com/google/uuid"
)

type Webhook struct {
	Game *Game  `json:"game"`
	Id   int    `json:"id"`
	Uuid string `json:"uuid"`
}

const WebhookKeyLength = 36

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
		webhook := &Webhook{Game: game}

		err := rows.Scan(&webhook.Id, &webhook.Uuid)
		if err != nil {
			log.Printf("Unable to scan webhook: %v.\r\n", err)
			return err
		}

		game.webhooks[webhook.Uuid] = webhook
	}

	return nil
}

func (game *Game) DeleteWebhook(webhook *Webhook) error {
	result, err := game.db.Exec(`
	DELETE FROM
		webhooks
	WHERE
		id = ?`, webhook.Id)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	delete(game.webhooks, webhook.Uuid)
	return nil
}

func (webhook *Webhook) DetachScript(script *Script) error {
	result, err := webhook.Game.db.Exec(`
		DELETE FROM
			webhook_script
		WHERE
			webhook_id = ?
		AND
			script_id = ?`, webhook.Id, script.Id)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	delete(webhook.Game.webhookScripts, webhook.Id)
	return nil
}

func (webhook *Webhook) AttachScript(script *Script) error {
	_, err := webhook.Game.db.Exec(`
	INSERT INTO
		webhook_script(webhook_id, script_id)
	VALUES
		(?, ?)
	`, webhook.Id, script.Id)
	if err != nil {
		return err
	}

	webhook.Game.webhookScripts[webhook.Id] = script
	return nil
}

func (game *Game) CreateWebhook() (*Webhook, error) {
	webhookUuid, err := uuid.NewUUID()
	if err != nil {
		return nil, err
	}

	uuidString := webhookUuid.String()

	/* Try to create the new webhook in-DB first */
	res, err := game.db.Exec(`
	INSERT INTO
		webhooks(uuid)
	VALUES
		(?)
	`, uuidString)
	if err != nil {
		return nil, err
	}

	var insertId int64

	insertId, err = res.LastInsertId()
	if err != nil {
		return nil, err
	}

	game.webhooks[uuidString] = &Webhook{Id: int(insertId), Uuid: uuidString, Game: game}
	return game.webhooks[uuidString], nil
}

func (game *Game) handleWebhooks() {
	defer func() {
		recover()
	}()

	http.HandleFunc("/webhook", func(w http.ResponseWriter, req *http.Request) {
		keyParam := req.URL.Query().Get("key")

		if len(keyParam) != WebhookKeyLength {
			log.Print("Ignoring a webhook key submitted without a length of 36.\r\n")
			return
		}

		log.Printf("Got a webhook request with key: %s\r\n", keyParam)
		game.webhookMessage <- keyParam
	})

	http.ListenAndServe(":9000", nil)
}
