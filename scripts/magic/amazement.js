/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_amazement(ch) {
    ch.send("This spell is not currently implemented, please try again later!\r\n");
}

Golem.registerSpellHandler('amazement', spell_amazement);
