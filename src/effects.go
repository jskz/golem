/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import "github.com/dop251/goja"

type Effect struct {
	handler  goja.Callable
	callback goja.Callable
}

func (game *Game) createEffect(fn goja.Callable) goja.Value {
	result, err := fn(game.vm.ToValue(game), nil)
	if err != nil {
		return game.vm.ToValue(nil)
	}

	f, res := goja.AssertFunction(result)
	if !res {
		return game.vm.ToValue(nil)
	}

	effect := &Effect{}
	effect.handler = fn
	effect.callback = f

	return game.vm.ToValue(nil)
}
