/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_cure_light(ch, args) {
    const target = args.length > 1 ? ch.findCharacterInRoom(args) : ch;

    if (!target || !ch.room || !target.room || !target.room.isEqual(ch.room)) {
        ch.send("Your target isn't here.\r\n");
        return;
    }

    const amount = ~~(((Math.random() * 5) + 5) * (this.proficiency / 100));

    Golem.game.damage(null, target, false, -amount, Golem.Combat.DamageTypeExotic);
    target.send('{WYou feel a little bit better.{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(target)) {
            rch.send(
                '{W' +
                    target.getShortDescriptionUpper(rch) +
                    ' looks a little bit better.{x\r\n'
            );
        }
    }
}

Golem.registerSpellHandler('cure light', spell_cure_light);
