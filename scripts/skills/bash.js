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
    if (!victim) {
        ch.send('Bash who?\r\n');
        return;
    }

    if(ch.stamina < 25) {
        ch.send("You are too tired to do that.\r\n");
        return;
    }

    ch.stamina = Math.max(0, ch.stamina - 25);
    ch.send(
        'You bash into ' +
            victim.getShortDescription(ch) +
            ', sending them flying!\r\n'
    );

    const amount = ~~(((Math.random() * ch.level) * 1.5) * (this.proficiency / 100));
    Golem.game.damage(ch, victim, false, amount, Golem.Combat.DamageTypeBash);
}

Golem.registerSkillHandler('bash', do_bash);
