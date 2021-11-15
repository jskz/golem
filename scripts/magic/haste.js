/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_haste(ch, args) {
    const target = args.length > 1 ? ch.findCharacterInRoom(args) : ch;

    if(target.affected & Golem.AffectedTypes.AFFECT_HASTE) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    target.addEffect(Golem.game.createEffect(Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_HASTE,
        ch.level * 2, // duration
        ch.level,
        0,
        0,
        function() {
            target.send("{DYou slow down and begin to move normally again.{x\r\n");

            for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
                const rch = iter.value;

                if (!rch.isEqual(target)) {
                    rch.send(
                        '{D' +
                            target.getShortDescriptionUpper(rch) +
                            ' slows down and begins to move normally again.{x\r\n'
                    );
                }
            }
        }));

    target.send('{DYou accelerate and your movements begin to blur.{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(target)) {
            rch.send(
                '{D' +
                    target.getShortDescriptionUpper(rch) +
                    ' begins to move and blur with impossible agility.{x\r\n'
            );
        }
    }
}

Golem.registerSpellHandler('haste', spell_haste);
