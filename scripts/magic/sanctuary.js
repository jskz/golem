/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_sanctuary(ch) {
    if(ch.affected & Golem.AffectedTypes.AFFECT_SANCTUARY) {
        ch.send("{WYou failed.{x\r\n");
        return;
    }

    ch.addEffect(Golem.game.createEffect(Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_SANCTUARY,
        ch.level * 2, // duration
        ch.level,
        0,
        0,
        function() {
            ch.send("{WYour holy protection has worn off.{x\r\n");
        }));

    ch.send("{WYou feel protected.{x\r\n");
}

Golem.registerSpellHandler('sanctuary', spell_sanctuary);
