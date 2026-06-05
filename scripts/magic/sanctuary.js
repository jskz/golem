/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_sanctuary(ch, args) {
    const target = args.length > 1 ? ch.findCharacterInRoom(args) : ch;

    if (!target || !ch.room || !target.room || !target.room.isEqual(ch.room)) {
        ch.send("Your target isn't here.\r\n");
        return;
    }

    if (target.affected & Golem.AffectedTypes.AFFECT_SANCTUARY) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    target.addEffect(Golem.game.createEffect(
        'sanctuary',
        Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_SANCTUARY,
        ch.level * 2, // duration
        ch.level,
        0,
        0,
        function(affected) {
            if (!affected) {
                return;
            }

            affected.send("{WYour holy protection has worn off.{x\r\n");

            if (!affected.room) {
                return;
            }

            for (let iter = affected.room.characters.head; iter !== null; iter = iter.next) {
                const rch = iter.value;

                if (!rch.isEqual(affected)) {
                    rch.send(
                        '{WThe protective aura surrounding ' +
                            affected.getShortDescription(rch) +
                            ' fades and va{wnishes.{x\r\n'
                    );
                }
            }
        }));

    target.send('{WYou feel protected.{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(target)) {
            rch.send(
                '{W' +
                    target.getShortDescriptionUpper(rch) +
                    ' is surrounded by a soft white aura.{x\r\n'
            );
        }
    }
}

Golem.registerSpellHandler('sanctuary', spell_sanctuary);
