/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_fireball(ch, args) {
    if (!args.length) {
        ch.send('This spell requires a target.\r\n');
        return;
    }

    const target = ch.findCharacterInRoom(args);

    if (!target || !ch.room || !target.room || target.room.id !== ch.room.id) {
        ch.send("Your target isn't here.\r\n");
        return;
    }

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(target)) {
            rch.send(
                '{R' +
                    target.getShortDescriptionUpper(rch) +
                    ' bursts into flames!{x\r\n'
            );
        }
    }

    target.send('{RYou are enveloped in flames!{x\r\n');
    Golem.game.damage(ch, target, false, 25, Golem.Combat.DamageTypeExotic);
}

Golem.registerSpellHandler('fireball', spell_fireball);
