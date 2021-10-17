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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dop251/goja"
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
	log.Printf(fmt.Sprintf("Loading scripts from directory %s:\r\n", directory))
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
			log.Printf(fmt.Sprintf("Loading script: %s\r\n", path))
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
	ch.game.InvokeNamedEventHandlersWithContextAndArguments("reload", ch.game.vm.ToValue(ch.game))

	err := ch.game.LoadScripts()
	if err != nil {
		ch.Send(fmt.Sprintf("{RFailed reload: %s{x\r\n", err.Error()))
		return
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
			skill.handler = nil
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

	roomFlagsConstantsObj := game.vm.NewObject()
	roomFlagsConstantsObj.Set("RoomPersistent", ROOM_PERSISTENT)
	roomFlagsConstantsObj.Set("RoomVirtual", ROOM_VIRTUAL)
	roomFlagsConstantsObj.Set("RoomSafe", ROOM_SAFE)

	combatObj := game.vm.NewObject()
	combatObj.Set("DamageTypeBash", game.vm.ToValue(DamageTypeBash))
	combatObj.Set("DamageTypeSlash", game.vm.ToValue(DamageTypeSlash))
	combatObj.Set("DamageTypeStab", game.vm.ToValue(DamageTypeStab))
	combatObj.Set("DamageTypeExotic", game.vm.ToValue(DamageTypeExotic))

	obj.Set("RoomFlags", roomFlagsConstantsObj)
	obj.Set("Combat", combatObj)

	game.vm.Set("Golem", obj)
	game.vm.Set("setTimeout", game.vm.ToValue(game.setTimeout))

	err := game.LoadScripts()
	if err != nil {
		return err
	}

	return nil
}
