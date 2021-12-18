/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_detect_magic(ch, args) {
    if(ch.affected & Golem.AffectedTypes.AFFECT_DETECT_MAGIC) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    ch.addEffect(Golem.game.createEffect('detect magic',
        Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_DETECT_MAGIC,
        ch.level * 6, // duration
        ch.level,
        0,
        0,
        function() {
            ch.send("{DYour pupils widen and your vision returns to normal.{x\r\n");
        }));

    ch.send('{DYour pupils dilate as brilliant leylines with the spirit world augment your vision.{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(ch)) {
            rch.send(
                '{D' +
                    ch.getShortDescriptionUpper(rch) +
                    ' suddenly shivers and momentarily stares placidly into space before snapping out of it.{x\r\n'
            );
        }
    }
}

Golem.registerSpellHandler('detect magic', spell_detect_magic);
