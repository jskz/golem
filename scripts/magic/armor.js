/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_armor(ch) {
    ch.defense += 5;
    ch.send("{WThe air suddenly hardens around you!{x\r\n");

    setTimeout(function() {
        ch.defense -= 5;
        ch.send("{DYour magical armor has worn off.{x\r\n");
    }, 5000);
}

Golem.registerSpellHandler('armor', spell_armor);