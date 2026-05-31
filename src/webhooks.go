/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
)

type Webhook struct {
	Game *Game  `json:"game"`
	Id   int    `json:"id"`
	Uuid string `json:"uuid"`
}

type WorldMapCharacterPointData struct {
	Name string `json:"name"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
}

type WorldMapResponse struct {
	Terrain    [][]int                      `json:"terrain"`
	Characters []WorldMapCharacterPointData `json:"characters"`
}

type worldMapRequest struct {
	response chan worldMapResult
}

type worldMapResult struct {
	response *WorldMapResponse
	err      error
}

const (
	WebhookKeyLength     = 36
	webhookListenAddress = ":9000"
)

var (
	errWorldMapOverworldNotFound    = errors.New("failed to find overworld plane")
	errWorldMapOverworldUnavailable = errors.New("failed to load overworld map")
)

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

	return rows.Err()
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
	delete(game.webhookScripts, webhook.Id)
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
	uuidString, err := newWebhookUUID()
	if err != nil {
		return nil, err
	}

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

func newWebhookUUID() (string, error) {
	uuidBytes := make([]byte, 16)
	_, err := rand.Read(uuidBytes)
	if err != nil {
		return "", err
	}

	uuidBytes[6] = (uuidBytes[6] & 0x0f) | 0x40
	uuidBytes[8] = (uuidBytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuidBytes[0:4],
		uuidBytes[4:6],
		uuidBytes[6:8],
		uuidBytes[8:10],
		uuidBytes[10:16],
	), nil
}

func (game *Game) handleWebhooks() {
	server := &http.Server{
		Addr:    webhookListenAddress,
		Handler: recoverHTTPPanics(game.webhookMux()),
	}

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("Webhook HTTP server failed: %v\r\n", err)
	}
}

func (game *Game) webhookMux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/worldmap", game.handleWorldMap)
	mux.HandleFunc("/webhook", game.handleWebhook)
	return mux
}

func recoverHTTPPanics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("Webhook HTTP handler panicked on %s %s: %v\r\n%s", req.Method, req.URL.Path, recovered, debug.Stack())
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, req)
	})
}

func (game *Game) handleWorldMap(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if game.worldMapRequest == nil {
		http.Error(w, "world map service is unavailable", http.StatusServiceUnavailable)
		return
	}

	request := worldMapRequest{response: make(chan worldMapResult, 1)}

	select {
	case game.worldMapRequest <- request:
	case <-req.Context().Done():
		return
	}

	var result worldMapResult

	select {
	case result = <-request.response:
	case <-req.Context().Done():
		return
	}

	if result.err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(result.err.Error()))
		return
	}

	encoded, err := json.Marshal(result.response)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(fmt.Sprintf("failed to encode overworld terrain: %v", err)))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(encoded)
}

func (game *Game) buildWorldMapResponse() (*WorldMapResponse, error) {
	overworld := game.FindPlaneByName("overworld")
	if overworld == nil {
		return nil, errWorldMapOverworldNotFound
	}

	if overworld.Map == nil || len(overworld.Map.Layers) == 0 || overworld.Map.Layers[0] == nil {
		return nil, errWorldMapOverworldUnavailable
	}

	layer := overworld.Map.Layers[0]
	response := &WorldMapResponse{
		Terrain:    copyWorldMapTerrain(layer.Terrain),
		Characters: make([]WorldMapCharacterPointData, 0),
	}

	if layer.Atlas == nil || layer.Atlas.CharacterTree == nil || layer.Atlas.CharacterTree.Boundary == nil {
		return response, nil
	}

	overworldCharacters := layer.Atlas.CharacterTree.QueryRect(layer.Atlas.CharacterTree.Boundary)
	response.Characters = make([]WorldMapCharacterPointData, 0, len(overworldCharacters))

	for _, ochPoint := range overworldCharacters {
		och, ok := ochPoint.Value.(*Character)
		if !ok || och == nil {
			continue
		}

		response.Characters = append(response.Characters, WorldMapCharacterPointData{
			X:    int(ochPoint.X),
			Y:    int(ochPoint.Y),
			Name: och.Name,
		})
	}

	return response, nil
}

func copyWorldMapTerrain(terrain [][]int) [][]int {
	copied := make([][]int, len(terrain))

	for y := range terrain {
		copied[y] = append([]int(nil), terrain[y]...)
	}

	return copied
}

func (game *Game) handleWebhook(w http.ResponseWriter, req *http.Request) {
	keyParam := req.URL.Query().Get("key")

	if len(keyParam) != WebhookKeyLength {
		log.Print("Ignoring a webhook key submitted without a length of 36.\r\n")
		return
	}

	log.Print("Got a webhook request with a valid-length key.\r\n")
	game.webhookMessage <- keyParam
}
