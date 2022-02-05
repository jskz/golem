/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_blindness(ch, args) {
    const target = ch.findCharacterInRoom(args);

    if(!target) {
        ch.send("They aren't here.\r\n");
        return;
    }

    if(target.affected & Golem.AffectedTypes.AFFECT_BLINDNESS) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    target.addEffect(Golem.game.createEffect(
        'blindness',
        Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_BLINDNESS,
        ~~(ch.level / 3), // duration
        ch.level,
        0,
        0,
        function() {
            target.send("{wYou are able to see your surroundings again.{x\r\n");
        }));

    target.send('{DYour vision is clouded by darkness.{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(target)) {
            rch.send(
                '{D' +
                    target.getShortDescriptionUpper(rch) +
                    ' {Rbegins flailing about, unable to see their surroundings.{x\r\n'
            );
        }
    }
}

Golem.registerSpellHandler('blindness', spell_blindness);
