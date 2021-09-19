/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_fireball(ch, args) {
    ch.send("Fireball spell!\r\n");
}

Golem.registerSpellHandler('fireball', spell_fireball);