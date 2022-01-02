/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja/parser"
	"github.com/getsentry/sentry-go"
)

type EventHandler struct {
	name     string
	callback goja.Callable
}

type ScriptTimer struct {
	createdAt time.Time
	callback  goja.Callable
	delay     int64
}

type Script struct {
	Game    *Game        `json:"game"`
	Id      uint         `json:"id"`
	Name    string       `json:"name"`
	Script  string       `json:"script"`
	Exports *goja.Object `json:"exports"`
}

func (game *Game) DefaultSourceLoader(filename string) ([]byte, error) {
	for _, script := range game.Scripts {
		if strings.Compare(strings.ToLower(filename), strings.ToLower(script.Name)) == 0 {
			return []byte(script.Script), nil
		}
	}

	return nil, errors.New("unable to find a script of that name")
}

func (script *Script) GetExports() (*goja.Object, error) {
	game := script.Game

	source := "(function(exports, require, module) {" + script.Script + "\n})"
	parsed, err := goja.Parse(script.Name, source, parser.WithSourceMapLoader(game.DefaultSourceLoader))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	compiled, err := goja.CompileAST(parsed, false)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	res, err := game.vm.RunProgram(compiled)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	fn, ok := goja.AssertFunction(res)
	if !ok {
		log.Printf("Failed to execute script %s/%d.", script.Name, script.Id)
		return nil, nil
	}

	module := game.vm.NewObject()
	exports := game.vm.NewObject()
	module.Set("exports", exports)

	_, err = fn(exports, exports, nil, module)
	if err != nil {
		log.Printf("Failed to evaluate script (%s/%d): %v\r\n", script.Name, script.Id, err)
		return nil, err
	}

	return module.ToObject(game.vm).Get("exports").ToObject(game.vm), nil
}

func (game *Game) LoadScriptsFromDatabase() error {
	game.Scripts = make(map[uint]*Script)
	game.objectScripts = make(map[uint]*Script)
	game.webhookScripts = make(map[int]*Script)
	game.districtScripts = make(map[int]*Script)

	rows, err := game.db.Query(`
		SELECT
			scripts.id,
			scripts.name,
			scripts.script
		FROM
			scripts
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		script := &Script{Game: game}

		err := rows.Scan(&script.Id, &script.Name, &script.Script)
		if err != nil {
			log.Printf("Failed to load script from database: %v\r\n", err)
			continue
		}

		script.Exports, err = script.GetExports()
		if err != nil {
			log.Printf("Failed to retrieve module exports: %v\r\n", err)
			continue
		}

		game.Scripts[script.Id] = script
	}

	log.Printf("Loaded %d scripts from database.\r\n", len(game.Scripts))

	log.Println("Loading object-script relations from database...")
	rows, err = game.db.Query(`
		SELECT
			object_script.object_id,
			object_script.script_id
		FROM
			object_script
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var objectId uint
		var scriptId uint

		err := rows.Scan(&objectId, &scriptId)
		if err != nil {
			return err
		}

		_, ok := game.Scripts[scriptId]
		if !ok {
			log.Printf("Trying to relate object with script")
			continue
		}

		game.objectScripts[objectId] = game.Scripts[scriptId]
	}

	log.Println("Loading room-script relations from database...")
	rows, err = game.db.Query(`
		SELECT
			room_script.room_id,
			room_script.script_id
		FROM
			room_script
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var roomId uint
		var scriptId uint

		err := rows.Scan(&roomId, &scriptId)
		if err != nil {
			return err
		}

		_, ok := game.Scripts[scriptId]
		if !ok {
			log.Printf("Trying to relate room with script")
			continue
		}

		room, err := game.LoadRoomIndex(roomId)
		if err != nil {
			return err
		}

		room.script = game.Scripts[scriptId]
	}

	log.Println("Loading plane-script relations from database...")
	rows, err = game.db.Query(`
		SELECT
			plane_script.plane_id,
			plane_script.script_id
		FROM
			plane_script
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var planeId int
		var scriptId uint

		err := rows.Scan(&planeId, &scriptId)
		if err != nil {
			return err
		}

		plane := game.FindPlaneByID(planeId)
		if plane == nil {
			return errors.New("tried to load a plane_script for a nonexistent plane")
		}

		_, ok := game.Scripts[scriptId]
		if !ok {
			return errors.New("tried to load a plane_script for a nonexistent script")
		}

		plane.Scripts = game.Scripts[scriptId]
	}

	log.Println("Loading webhook-script relations from database...")
	rows, err = game.db.Query(`
		SELECT
			webhook_script.webhook_id,
			webhook_script.script_id
		FROM
			webhook_script
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var webhookId int
		var scriptId uint

		err := rows.Scan(&webhookId, &scriptId)
		if err != nil {
			return err
		}

		_, ok := game.Scripts[scriptId]
		if !ok {
			return errors.New("tried to load a webhook_script for a nonexistent script")
		}

		game.webhookScripts[webhookId] = game.Scripts[scriptId]
	}

	log.Println("Loading district-script relations from database...")
	rows, err = game.db.Query(`
		SELECT
			district_script.district_id,
			district_script.script_id
		FROM
			district_script
	`)
	if err != nil {
		return err
	}

	defer rows.Close()

	for rows.Next() {
		var districtId int
		var scriptId uint

		err := rows.Scan(&districtId, &scriptId)
		if err != nil {
			return err
		}

		_, ok := game.Scripts[scriptId]
		if !ok {
			return errors.New("tried to load a webhook_script for a nonexistent script")
		}

		game.districtScripts[districtId] = game.Scripts[scriptId]
	}

	return nil
}

func (game *Game) setTimeout(cb goja.Callable, delay int64) goja.Value {
	defer func() {
		recover()
	}()

	timer := &ScriptTimer{}
	timer.createdAt = time.Now()
	timer.callback = cb
	timer.delay = delay
	game.ScriptTimers.Insert(timer)

	return game.vm.ToValue(timer)
}

func (game *Game) DeleteScript(script *Script) error {
	result, err := game.db.Exec(`
	DELETE FROM
		scripts
	WHERE
		id = ?`, script.Id)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	if err != nil {
		return err
	}

	delete(game.Scripts, script.Id)
	return nil
}

func (script *Script) Save() bool {
	result, err := script.Game.db.Exec(`
		UPDATE
			scripts
		SET
			name = ?,
			script = ?
		WHERE
			id = ?
	`, script.Name, script.Script, script.Id)
	if err != nil {
		log.Printf("Failed to save script: %v.\r\n", err)
		return false
	}

	_, err = result.RowsAffected()
	if err != nil {
		log.Printf("Failed to retrieve number of rows affected: %v.\r\n", err)
		return false
	}

	return true
}

func (game *Game) CreateScript(name string, initialBody string) (*Script, error) {
	res, err := game.db.Exec(`
	INSERT INTO
		scripts(name, script)
	VALUES
		(?, ?)
	`, name, initialBody)
	if err != nil {
		return nil, err
	}

	insertId64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	insertId := uint(insertId64)
	script := &Script{Id: insertId, Name: name, Script: initialBody, Game: game}
	game.Scripts[insertId] = script
	return script, nil
}

func (game *Game) scriptTimersUpdate() {
	for iter := game.ScriptTimers.Head; iter != nil; iter = iter.Next {
		effect := iter.Value.(*ScriptTimer)

		if time.Since(effect.createdAt).Milliseconds() > effect.delay {
			effect.callback(game.vm.ToValue(effect))
			game.ScriptTimers.Remove(effect)
			break
		}
	}
}

func (script *Script) tryEvaluate(methodName string, this goja.Value, arguments ...goja.Value) (goja.Value, error) {
	v := script.Exports.Get(methodName)
	fn, ok := goja.AssertFunction(v)
	if !ok {
		return nil, fmt.Errorf("%s not a function exported by script %s", methodName, script.Name)
	}

	result, err := fn(this, arguments...)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (game *Game) InvokeNamedEventHandlersWithContextAndArguments(name string, this goja.Value, arguments ...goja.Value) ([]goja.Value, []error) {
	if game.eventHandlers[name] != nil {
		values := make([]goja.Value, game.eventHandlers[name].Count)
		errors := make([]error, game.eventHandlers[name].Count)
		i := 0

		for iter := game.eventHandlers[name].Head; iter != nil; iter = iter.Next {
			eventHandler := iter.Value.(*EventHandler)

			result, err := eventHandler.callback(this, arguments...)
			if err != nil {
				errors[i] = err
				values[i] = nil
				i++

				continue
			}

			values[i] = result
			errors[i] = nil
			i++
		}

		return values, errors
	}

	return nil, nil
}

func (game *Game) LoadScripts() error {
	const ScriptDirectory = "scripts"

	return game.LoadScriptsFromDirectory(ScriptDirectory)
}

func (game *Game) LoadScriptsFromDirectory(directory string) error {
	log.Printf("Loading scripts from directory %s:\r\n", directory)
	scripts, err := ioutil.ReadDir(directory)
	if err != nil {
		return err
	}

	for _, filename := range scripts {
		path := fmt.Sprintf("%s/%s", directory, filename.Name())

		file, err := os.Stat(path)
		if err != nil {
			return err
		}

		fileFlags := file.Mode()
		if fileFlags.IsDir() {
			err := game.LoadScriptsFromDirectory(fmt.Sprintf("%s/%s", directory, filename.Name()))
			if err != nil {
				return err
			}
		} else {
			log.Printf("Loading script: %s\r\n", path)
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			_, err = game.vm.RunString(string(bytes))
			if err != nil {
				log.Println(err)
				return err
			}
		}
	}

	return nil
}

func do_reload(ch *Character, arguments string) {
	ch.Game.InvokeNamedEventHandlersWithContextAndArguments("reload", ch.Game.vm.ToValue(ch.Game))

	err := ch.Game.LoadScripts()
	if err != nil {
		ch.Send(fmt.Sprintf("{RFailed reload: %s{x\r\n", err.Error()))
		return
	}

	err = ch.Game.LoadScriptsFromDatabase()
	if err != nil {
		ch.Send(fmt.Sprintf("{RFailed database reload: %s{x\r\n", err.Error()))
	}

	ch.Send("{GScripts reloaded.{x\r\n")
}

func (game *Game) InitScripting() error {
	game.vm = goja.New()
	game.eventHandlers = make(map[string]*LinkedList)

	game.vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	obj := game.vm.NewObject()

	obj.Set("game", game.vm.ToValue(game))

	obj.Set("clearAllEventHandlers", game.vm.ToValue(func() goja.Value {
		game.eventHandlers = make(map[string]*LinkedList)

		return game.vm.ToValue(true)
	}))

	obj.Set("clearScriptedCommandHandlers", game.vm.ToValue(func() goja.Value {
		for _, cmd := range CommandTable {
			cmd.Callback = nil
		}

		return game.vm.ToValue(true)
	}))

	obj.Set("clearScriptedSkillHandlers", game.vm.ToValue(func() goja.Value {
		for _, skill := range game.skills {
			skill.Handler = nil
		}

		return game.vm.ToValue(true)
	}))

	obj.Set("registerEventHandler", game.vm.ToValue(func(name goja.Value, fn goja.Callable) goja.Value {
		eventName := name.String()
		if game.eventHandlers[eventName] == nil {
			game.eventHandlers[eventName] = NewLinkedList()
		}

		handler := &EventHandler{name: eventName, callback: fn}
		game.eventHandlers[eventName].Insert(handler)

		return game.vm.ToValue(handler)
	}))

	obj.Set("registerSkillHandler", game.vm.ToValue(func(name goja.Value, fn goja.Callable) goja.Value {
		skillName := name.String()

		return game.vm.ToValue(game.RegisterSkillHandler(skillName, fn))
	}))

	obj.Set("registerSpellHandler", game.vm.ToValue(func(name goja.Value, fn goja.Callable) goja.Value {
		spellName := name.String()

		return game.vm.ToValue(game.RegisterSpellHandler(spellName, fn))
	}))

	obj.Set("registerPlayerCommand", game.vm.ToValue(func(name goja.Value, fn goja.Callable) goja.Value {
		command := strings.ToLower(name.String())
		scriptedCommand := Command{
			Name:         command,
			Scripted:     true,
			CmdFunc:      nil,
			MinimumLevel: 0,
			Callback:     fn,
		}

		CommandTable[command] = scriptedCommand

		return game.vm.ToValue(scriptedCommand)
	}))

	knownLocationsConstantsObj := game.vm.NewObject()
	knownLocationsConstantsObj.Set("Limbo", RoomLimbo)
	knownLocationsConstantsObj.Set("DeveloperLounge", RoomDeveloperLounge)

	objectFlagsConstantsObj := game.vm.NewObject()
	objectFlagsConstantsObj.Set("ITEM_TAKE", ITEM_TAKE)
	objectFlagsConstantsObj.Set("ITEM_WEAPON", ITEM_WEAPON)
	objectFlagsConstantsObj.Set("ITEM_WEARABLE", ITEM_WEARABLE)
	objectFlagsConstantsObj.Set("ITEM_DECAYS", ITEM_DECAYS)
	objectFlagsConstantsObj.Set("ITEM_DECAY_SILENTLY", ITEM_DECAY_SILENTLY)
	objectFlagsConstantsObj.Set("ITEM_WEAR_HELD", ITEM_WEAR_HELD)
	objectFlagsConstantsObj.Set("ITEM_WEAR_TORSO", ITEM_WEAR_TORSO)
	objectFlagsConstantsObj.Set("ITEM_WEAR_BODY", ITEM_WEAR_BODY)
	objectFlagsConstantsObj.Set("ITEM_WEAR_NECK", ITEM_WEAR_NECK)
	objectFlagsConstantsObj.Set("ITEM_WEAR_LEGS", ITEM_WEAR_LEGS)
	objectFlagsConstantsObj.Set("ITEM_WEAR_HANDS", ITEM_WEAR_HANDS)
	objectFlagsConstantsObj.Set("ITEM_GLOW", ITEM_GLOW)
	objectFlagsConstantsObj.Set("ITEM_HUM", ITEM_HUM)
	objectFlagsConstantsObj.Set("ITEM_CLOSED", ITEM_CLOSED)
	objectFlagsConstantsObj.Set("ITEM_CLOSEABLE", ITEM_CLOSEABLE)
	objectFlagsConstantsObj.Set("ITEM_LOCKED", ITEM_LOCKED)
	objectFlagsConstantsObj.Set("ITEM_PERSISTENT", ITEM_PERSISTENT)

	charFlagsConstantsObj := game.vm.NewObject()
	charFlagsConstantsObj.Set("CHAR_AGGRESSIVE", CHAR_AGGRESSIVE)
	charFlagsConstantsObj.Set("CHAR_PRACTICE", CHAR_PRACTICE)
	charFlagsConstantsObj.Set("CHAR_IS_PLAYER", CHAR_IS_PLAYER)

	roomFlagsConstantsObj := game.vm.NewObject()
	roomFlagsConstantsObj.Set("ROOM_PERSISTENT", ROOM_PERSISTENT)
	roomFlagsConstantsObj.Set("ROOM_VIRTUAL", ROOM_VIRTUAL)
	roomFlagsConstantsObj.Set("ROOM_PLANAR", ROOM_PLANAR)
	roomFlagsConstantsObj.Set("ROOM_SAFE", ROOM_SAFE)
	roomFlagsConstantsObj.Set("ROOM_DUNGEON", ROOM_DUNGEON)
	roomFlagsConstantsObj.Set("ROOM_EVIL_AURA", ROOM_EVIL_AURA)

	exitFlagsConstantsObj := game.vm.NewObject()
	exitFlagsConstantsObj.Set("EXIT_IS_DOOR", EXIT_IS_DOOR)
	exitFlagsConstantsObj.Set("EXIT_CLOSED", EXIT_CLOSED)
	exitFlagsConstantsObj.Set("EXIT_LOCKED", EXIT_LOCKED)

	directionsConstantsObj := game.vm.NewObject()
	directionsConstantsObj.Set("DirectionNorth", DirectionNorth)
	directionsConstantsObj.Set("DirectionEast", DirectionEast)
	directionsConstantsObj.Set("DirectionSouth", DirectionSouth)
	directionsConstantsObj.Set("DirectionWest", DirectionWest)
	directionsConstantsObj.Set("DirectionUp", DirectionUp)
	directionsConstantsObj.Set("DirectionDown", DirectionDown)

	wearLocationsConstantsObj := game.vm.NewObject()
	wearLocationsConstantsObj.Set("WearLocationNone", game.vm.ToValue(WearLocationNone))
	wearLocationsConstantsObj.Set("WearLocationHead", game.vm.ToValue(WearLocationHead))
	wearLocationsConstantsObj.Set("WearLocationNeck", game.vm.ToValue(WearLocationNeck))
	wearLocationsConstantsObj.Set("WearLocationArms", game.vm.ToValue(WearLocationArms))
	wearLocationsConstantsObj.Set("WearLocationTorso", game.vm.ToValue(WearLocationTorso))
	wearLocationsConstantsObj.Set("WearLocationHands", game.vm.ToValue(WearLocationHands))
	wearLocationsConstantsObj.Set("WearLocationShield", game.vm.ToValue(WearLocationShield))
	wearLocationsConstantsObj.Set("WearLocationBody", game.vm.ToValue(WearLocationBody))
	wearLocationsConstantsObj.Set("WearLocationWaist", game.vm.ToValue(WearLocationWaist))
	wearLocationsConstantsObj.Set("WearLocationLegs", game.vm.ToValue(WearLocationLegs))
	wearLocationsConstantsObj.Set("WearLocationFeet", game.vm.ToValue(WearLocationFeet))
	wearLocationsConstantsObj.Set("WearLocationWielded", game.vm.ToValue(WearLocationWielded))
	wearLocationsConstantsObj.Set("WearLocationHeld", game.vm.ToValue(WearLocationHeld))
	wearLocationsConstantsObj.Set("WearLocationMax", game.vm.ToValue(WearLocationMax))

	effectTypes := game.vm.NewObject()
	effectTypes.Set("EffectTypeAffected", game.vm.ToValue(EffectTypeAffected))
	effectTypes.Set("EffectTypeStat", game.vm.ToValue(EffectTypeStat))
	effectTypes.Set("EffectTypeImmunity", game.vm.ToValue(EffectTypeImmunity))

	affectedTypes := game.vm.NewObject()
	affectedTypes.Set("AFFECT_SANCTUARY", game.vm.ToValue(AFFECT_SANCTUARY))
	affectedTypes.Set("AFFECT_HASTE", game.vm.ToValue(AFFECT_HASTE))
	affectedTypes.Set("AFFECT_SLOW", game.vm.ToValue(AFFECT_SLOW))
	affectedTypes.Set("AFFECT_POISON", game.vm.ToValue(AFFECT_POISON))
	affectedTypes.Set("AFFECT_SILENCE", game.vm.ToValue(AFFECT_SILENCE))
	affectedTypes.Set("AFFECT_DETECT_MAGIC", game.vm.ToValue(AFFECT_DETECT_MAGIC))
	affectedTypes.Set("AFFECT_FIRESHIELD", game.vm.ToValue(AFFECT_FIRESHIELD))
	affectedTypes.Set("AFFECT_PARALYSIS", game.vm.ToValue(AFFECT_PARALYSIS))

	statTypes := game.vm.NewObject()
	statTypes.Set("STAT_NONE", game.vm.ToValue(STAT_NONE))
	statTypes.Set("STAT_STRENGTH", game.vm.ToValue(STAT_STRENGTH))
	statTypes.Set("STAT_DEXTERITY", game.vm.ToValue(STAT_DEXTERITY))
	statTypes.Set("STAT_CONSTITUTION", game.vm.ToValue(STAT_CONSTITUTION))
	statTypes.Set("STAT_INTELLIGENCE", game.vm.ToValue(STAT_INTELLIGENCE))
	statTypes.Set("STAT_WISDOM", game.vm.ToValue(STAT_WISDOM))
	statTypes.Set("STAT_CHARISMA", game.vm.ToValue(STAT_CHARISMA))
	statTypes.Set("STAT_LUCK", game.vm.ToValue(STAT_LUCK))
	statTypes.Set("STAT_MAX", game.vm.ToValue(STAT_MAX))

	combatObj := game.vm.NewObject()
	combatObj.Set("DamageTypeBash", game.vm.ToValue(DamageTypeBash))
	combatObj.Set("DamageTypeSlash", game.vm.ToValue(DamageTypeSlash))
	combatObj.Set("DamageTypeStab", game.vm.ToValue(DamageTypeStab))
	combatObj.Set("DamageTypeExotic", game.vm.ToValue(DamageTypeExotic))

	httpUtilityObj := game.vm.NewObject()
	httpUtilityObj.Set("Get", game.vm.ToValue(SimpleGET))
	httpUtilityObj.Set("Post", game.vm.ToValue(SimplePOST))

	terrainTypes := game.vm.NewObject()
	terrainTypes.Set("TerrainTypeOcean", game.vm.ToValue(TerrainTypeOcean))
	terrainTypes.Set("TerrainTypeCaveTunnel", game.vm.ToValue(TerrainTypeCaveTunnel))
	terrainTypes.Set("TerrainTypeCaveDeepWall1", game.vm.ToValue(TerrainTypeCaveDeepWall1))
	terrainTypes.Set("OverworldCityExterior", game.vm.ToValue(TerrainTypeOverworldCityExterior))
	terrainTypes.Set("OverworldCityInterior", game.vm.ToValue(TerrainTypeOverworldCityInterior))
	terrainTypes.Set("OverworldCityEntrance", game.vm.ToValue(TerrainTypeOverworldCityEntrance))
	terrainTypes.Set("TerrainTypePlains", game.vm.ToValue(TerrainTypePlains))
	terrainTypes.Set("TerrainTypeField", game.vm.ToValue(TerrainTypeField))
	terrainTypes.Set("TerrainTypeShore", game.vm.ToValue(TerrainTypeShore))
	terrainTypes.Set("TerrainTypeShallowWater", game.vm.ToValue(TerrainTypeShallowWater))
	terrainTypes.Set("TerrainTypeLightForest", game.vm.ToValue(TerrainTypeLightForest))
	terrainTypes.Set("TerrainTypeDenseForest", game.vm.ToValue(TerrainTypeDenseForest))
	terrainTypes.Set("TerrainTypeHills", game.vm.ToValue(TerrainTypeHills))
	terrainTypes.Set("TerrainTypeMountains", game.vm.ToValue(TerrainTypeMountains))
	terrainTypes.Set("TerrainTypeSnowcappedMountains", game.vm.ToValue(TerrainTypeSnowcappedMountains))

	utilObj := game.vm.NewObject()
	utilObj.Set("createLinkedList", game.vm.ToValue(NewLinkedList))
	utilObj.Set("createQuadTree", game.vm.ToValue(NewQuadTree))
	utilObj.Set("newRect2D", game.vm.ToValue(NewRect))
	utilObj.Set("newPoint2D", game.vm.ToValue(NewPoint))

	utilObj.Set("angle2D", game.vm.ToValue(Angle2D))
	utilObj.Set("distance2D", game.vm.ToValue(Distance2D))
	utilObj.Set("perlin2D", game.vm.ToValue(Perlin2D))
	utilObj.Set("angleToDirection", game.vm.ToValue(AngleToDirection))
	utilObj.Set("oneArgument", game.vm.ToValue(OneArgument))

	obj.Set("KnownLocations", knownLocationsConstantsObj)
	obj.Set("ExitFlags", exitFlagsConstantsObj)
	obj.Set("ExitName", game.vm.ToValue(ExitName))
	obj.Set("EffectTypes", effectTypes)
	obj.Set("AffectedTypes", affectedTypes)
	obj.Set("RoomFlags", roomFlagsConstantsObj)
	obj.Set("TerrainTypes", terrainTypes)
	obj.Set("StatTypes", statTypes)
	obj.Set("CharacterFlags", charFlagsConstantsObj)
	obj.Set("ObjectFlags", objectFlagsConstantsObj)
	obj.Set("Combat", combatObj)
	obj.Set("Directions", directionsConstantsObj)
	obj.Set("WearLocations", wearLocationsConstantsObj)
	obj.Set("HTTP", httpUtilityObj)
	obj.Set("NewExit", game.vm.ToValue(game.NewExit))

	obj.Set("util", utilObj)

	sentryObj := game.vm.NewObject()
	sentryObj.Set("captureMessage", game.vm.ToValue(sentry.CaptureMessage))

	game.vm.Set("Golem", obj)
	game.vm.Set("println", game.vm.ToValue(log.Println))
	game.vm.Set("Sentry", sentryObj)
	game.vm.Set("setTimeout", game.vm.ToValue(game.setTimeout))

	err := game.LoadScripts()
	if err != nil {
		return err
	}

	err = game.LoadScriptsFromDatabase()
	if err != nil {
		return err
	}

	return nil
}
