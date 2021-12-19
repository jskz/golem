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

    const weapon = ch.getEquipment(Golem.WearLocations.WearLocationWielded);
    if(!weapon) {
        ch.send("You can't backstab without a weapon.\r\n");
        return;
    }

    const damageType = parseInt(weapon.value3);
    if(damageType !== Golem.Combat.DamageTypeStab) {
        ch.send(weapon.getShortDescriptionUpper() + "{x isn't a stabbing weapon.\r\n");
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
     * Sanctuary protects against instant-kill scenarios, so is checked here.
     *
     * In this "artful edge case", the victim will not receive a message - only the death ANSI.
     */
    if((this.proficiency === 100
    && ~~(Math.random() * 100) > 98
    && victim.level < ch.level
    && !(victim.affected & Golem.AffectedTypes.AFFECT_SANCTUARY))) {
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
                        '{D steps behind ' + victim.getShortDescription(rch) + '{D and pierces their heart!{x\r\n'
                );
            }
        }

        Golem.game.damage(ch, victim, false, victim.health, Golem.Combat.DamageTypeStab);
        ch.client.delay(2000);
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
    ch.client.delay(2000);
}

Golem.registerSkillHandler('backstab', do_backstab);
