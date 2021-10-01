/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
package main

import (
	"time"

	"github.com/dop251/goja"
)

type Effect struct {
	createdAt time.Time
	handler   goja.Callable
	callback  goja.Callable
	delay     int64
}

func (game *Game) createEffect(fn goja.Callable) goja.Value {
	defer func() {
		recover()
	}()

	result, err := fn(game.vm.ToValue(game), nil)
	if err != nil {
		return game.vm.ToValue(nil)
	}

	obj := result.ToObject(game.vm)
	cb := obj.Get("0")
	delay := obj.Get("1").ToInteger()

	f, res := goja.AssertFunction(cb)
	if !res {
		return game.vm.ToValue(nil)
	}

	effect := &Effect{}
	effect.createdAt = time.Now()
	effect.handler = fn
	effect.callback = f
	effect.delay = delay

	game.Effects.Insert(effect)

	return game.vm.ToValue(effect)
}

func (game *Game) effectsUpdate() {
	for iter := game.Effects.Head; iter != nil; iter = iter.Next {
		effect := iter.Value.(*Effect)

		if time.Since(effect.createdAt).Milliseconds() > effect.delay {
			effect.callback(game.vm.ToValue(effect))
			game.Effects.Remove(effect)
			break
		}
	}
}
