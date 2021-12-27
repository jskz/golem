/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */

// moderate healing for anybody in the caster's party, including the caster
function spell_group_heal(ch) {
    if (!ch.room) {
        ch.send("You can't cast that here.\r\n");
        return;
    }

    // re-use same amount roll for every target
    const amount = ~~(((Math.random() * (ch.level * 4)) + ch.level) * (this.proficiency / 100));

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.inSameGroup(ch)) {
            continue;
        }

        Golem.game.damage(null, rch, false, -amount, Golem.Combat.DamageTypeExotic);
        rch.send('{WYou feel better.{x\r\n');

        for (let innerIter = ch.room.characters.head; innerIter !== null; innerIter = innerIter.next) {
            const och = innerIter.value;

            if (!och.isEqual(rch)) {
                och.send(
                    '{W' + rch.getShortDescriptionUpper(och)
                    + ' looks better.{x\r\n'
                );
            }
        }
    }
}

Golem.registerSpellHandler('group heal', spell_group_heal);
