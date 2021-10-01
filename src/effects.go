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
	callback  goja.Callable
	delay     int64
}

func (game *Game) createEffect(cb goja.Callable, delay int64) goja.Value {
	defer func() {
		recover()
	}()

	effect := &Effect{}
	effect.createdAt = time.Now()
	effect.callback = cb
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
