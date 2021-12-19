/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_stun(ch, args) {
    let victim =
        ch.fighting !== null ? ch.fighting : ch.findCharacterInRoom(args);
    if (!victim || victim.isEqual(ch)) {
        ch.send('Stun who?\r\n');
        return;
    }

    if(ch.stamina < 75) {
        ch.send("You are too tired to do that.\r\n");
        return;
    }

    ch.stamina = Math.max(0, ch.stamina - 75);
    ch.send(
        '{YYou smash into ' +
            victim.getShortDescription(ch) +
            '{Y, leaving them stunned!{x\r\n'
    );

    victim.send('{Y' + ch.getShortDescriptionUpper(victim) + '{Y smashes into you, leaving you stunned!{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(victim) && !rch.isEqual(ch)) {
            rch.send(
                '{Y' + ch.getShortDescriptionUpper(rch) +
                    '{Y smashes into ' + victim.getShortDescription(rch) + '{Y, leaving them stunned!{x\r\n'
            );
        }
    }

    victim.addEffect(Golem.game.createEffect(
        'paralysis',
        Golem.EffectTypes.EffectTypeAffected,
        Golem.AffectedTypes.AFFECT_PARALYSIS,
        ~~(Math.random() * (ch.level / 15)) + 1,
        ch.level,
        0,
        0,
        function() {
            victim.send("{YYour senses recover and you are no longer stunned.{x\r\n");

            for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
                const rch = iter.value;

                if (!rch.isEqual(victim)) {
                    rch.send(
                        '{Y' + victim.getShortDescriptionUpper(rch) +
                            '{Y returns to their senses.{x\r\n'
                    );
                }
            }
    }));

    ch.client.delay(4000);
}

Golem.registerSkillHandler('stun', do_stun);
