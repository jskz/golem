/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_sanctuary(ch, args) {
    const target = args.length > 1 ? ch.findCharacterInRoom(args) : ch;

    if(target.affected & Golem.AffectedTypes.AFFECT_SANCTUARY) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    target.addEffect(Golem.game.createEffect(Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_SANCTUARY,
        ch.level * 2, // duration
        ch.level,
        0,
        0,
        function() {
            ch.send("{WYour holy protection has worn off.{x\r\n");

            for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
                const rch = iter.value;

                if (!rch.isEqual(target)) {
                    rch.send(
                        '{WThe protective aura surrounding ' +
                            target.getShortDescription(rch) +
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
