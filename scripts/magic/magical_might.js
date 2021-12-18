/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_magical_might(ch, args) {
    if(ch.affected & Golem.AffectedTypes.AFFECT_MAGICAL_MIGHT) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    ch.addEffect(Golem.game.createEffect(
        'magical might',
        Golem.EffectTypes.EffectTypeStat,
        0,
        ch.level * 6, // duration
        ch.level,
        Golem.StatTypes.STAT_STRENGTH,
        3,
        function() {
            ch.send("{DThe magical energy pulsing through your muscles subsides.{x\r\n");
        }));

    ch.send('{MYour muscles harden as a magical energy surges through you.{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(ch)) {
            rch.send(
                '{M' +
                    ch.getShortDescriptionUpper(rch) +
                    '{M is strengthened by a magical surge.{x\r\n'
            );
        }
    }
}

Golem.registerSpellHandler('magical might', spell_magical_might);
