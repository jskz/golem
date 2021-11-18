/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_fireshield(ch, args) {
    const target = args.length > 1 ? ch.findCharacterInRoom(args) : ch;

    if(target.affected & Golem.AffectedTypes.AFFECT_FIRESHIELD) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    target.addEffect(Golem.game.createEffect(Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_FIRESHIELD,
        ch.level, // duration
        ch.level,
        0,
        0,
        function() {
            target.send("{RYour reactive fireshield vanishes.{x\r\n");

            for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
                const rch = iter.value;

                if (!rch.isEqual(target)) {
                    rch.send(
                        '{RThe reactive fireshield surrounding ' +
                            target.getShortDescription(rch) +
                            ' {Rsputters and dies.{x\r\n'
                    );
                }
            }
        }));

    target.send('{RYou are surrounding by a crackling fireshield.{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(target)) {
            rch.send(
                '{R' +
                    target.getShortDescriptionUpper(rch) +
                    ' {Ris enveloped by an intense, crackling fireshield.{x\r\n'
            );
        }
    }
}

Golem.registerSpellHandler('fireshield', spell_fireshield);
