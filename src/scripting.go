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
	"strings"

	"github.com/dop251/goja"
)

func (game *Game) InitScripting() error {
	game.vm = goja.New()

	game.vm.SetFieldNameMapper(goja.TagFieldNameMapper("json", true))

	obj := game.vm.NewObject()
	obj.Set("registerPlayerCommand", game.vm.ToValue(func(name goja.Value, fn goja.Callable) goja.Value {
		command := strings.ToLower(name.String())
		_, ok := CommandTable[command]
		if ok {
			log.Printf("Trying to register duplicate command, aborting.\r\n")
			return game.vm.ToValue(false)
		}

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

	game.vm.Set("Golem", obj)

	return nil
}
