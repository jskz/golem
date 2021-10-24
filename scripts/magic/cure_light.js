/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_cure_light(ch, args) {
    const amount = ~~(Math.random() * 5) + 5;

    Golem.game.damage(null, ch, false, -amount, Golem.Combat.DamageTypeExotic);
    ch.send('{WYou feel a little bit better.{x\r\n');
}

Golem.registerSpellHandler('cure light', spell_cure_light);
