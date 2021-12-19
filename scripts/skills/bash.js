/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_bash(ch, args) {
    let victim =
        ch.fighting !== null ? ch.fighting : ch.findCharacterInRoom(args);
    if (!victim || victim.isEqual(ch)) {
        ch.send('Bash who?\r\n');
        return;
    }

    if(ch.stamina < 25) {
        ch.send("You are too tired to do that.\r\n");
        return;
    }

    ch.stamina = Math.max(0, ch.stamina - 25);
    ch.send(
        '{RYou bash into ' +
            victim.getShortDescription(ch) +
            '{R, sending them flying!{x\r\n'
    );

    victim.send('{R' + ch.getShortDescriptionUpper(victim) + '{R bashes into you, sending you flying!{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(victim) && !rch.isEqual(ch)) {
            rch.send(
                '{R' + ch.getShortDescriptionUpper(rch) +
                    '{R bashes into ' + victim.getShortDescriptionUpper(rch) + '{R, sending them flying!{x\r\n'
            );
        }
    }

    const amount = ~~(((Math.random() * ch.level) * 1.5) * (this.proficiency / 100));
    Golem.game.damage(ch, victim, false, amount, Golem.Combat.DamageTypeBash);
    ch.client.delay(2000);
}

Golem.registerSkillHandler('bash', do_bash);
