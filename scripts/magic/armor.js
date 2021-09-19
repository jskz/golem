/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_armor(ch, args) {
    ch.send("Armor spell!\r\n");
}

Golem.registerSpellHandler('armor', spell_armor);