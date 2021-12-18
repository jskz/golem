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

    /*
     * If the skill user is 100% proficient with backstab and the victim is of a level less
     * than the skill user, then allow for a 2% chance to instant-kill the target.
     *
     * In this "artful edge case", the victim will not receive a message - only the death ANSI.
     */
    if((ch.proficiency === 100
    && ~~(Math.random() * 100) > 98
    && victim.level < ch.level)) {
        ch.send(
            '{DYou artfully stab ' +
                victim.getShortDescription(ch) +
                '{D through the heart, killing them in cold blood!{x\r\n'
        );

        for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
            const rch = iter.value;

            if (!rch.isEqual(victim) && !rch.isEqual(ch)) {
                rch.send(
                    '{D' + ch.getShortDescriptionUpper(rch) +
                        '{D steps behind ' + victim.getShortDescription(rch) + '{D and pierces their heart, from behind!{x\r\n'
                );
            }
        }

        Golem.game.damage(ch, victim, false, victim.health, Golem.Combat.DamageTypeStab);
        return;
    }

    ch.send(
        '{RYou carefully approach ' +
            victim.getShortDescription(ch) +
            ' from behind and stab them in the back!{x\r\n'
    );

    victim.send('{RYou feel a terrible pain as ' + ch.getShortDescription(victim) + '{R stabs you in the back!{x\r\n');

    for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
        const rch = iter.value;

        if (!rch.isEqual(victim) && !rch.isEqual(ch)) {
            rch.send(
                '{R' + ch.getShortDescriptionUpper(rch) +
                    '{R steps behind ' + victim.getShortDescription(rch) + '{D and stabs them in the back!{x\r\n'
            );
        }
    }

    const amount = ~~(((Math.random() * ch.level) * 5) * (this.proficiency / 100));
    Golem.game.damage(ch, victim, false, amount, Golem.Combat.DamageTypeStab);
}

Golem.registerSkillHandler('backstab', do_backstab);
