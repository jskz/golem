/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_heal(ch) {
    ch.send("Heal spell!\r\n");
}

Golem.registerSpellHandler('heal', spell_heal);