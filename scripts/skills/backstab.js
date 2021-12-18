/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_backstab(ch, args) {
    if(ch.fighting || ch.combat) {
        ch.send("You can't do that while already fighting.\r\n");
        return;
    }

    let victim = ch.findCharacterInRoom(args);

    if (!victim) {
        ch.send('Backstab who?\r\n');
        return;
    }

    ch.send(
        '{RYou carefully approach ' +
            victim.getShortDescription(ch) +
            ' from behind and stab them in the back!{x\r\n'
    );

    const amount = ~~(((Math.random() * ch.level) * 5) * (this.proficiency / 100));
    Golem.game.damage(ch, victim, false, amount, Golem.Combat.DamageTypeStab);
}

Golem.registerSkillHandler('backstab', do_backstab);
