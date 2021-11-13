/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function onCombatUpdate() {
    for (let iter = this.fights.head; iter != null; iter = iter.next) {
        const combat = iter.value;

        let found = false;

        if(combat && combat.participants) {
            for (let i = 0; i < combat.participants.length; i++) {
                const vch = combat.participants[i];

                if(!vch.room) {
                    continue;
                }

                let attackerRounds = 1,
                    dexterityBonusRounds = parseInt((vch.dexterity - 10) / 4);

                attackerRounds += dexterityBonusRounds;

                for (let r = 0; r < attackerRounds; r++) {
                    let victim = vch.fighting;

                    if (
                        !victim ||
                        victim.room === null ||
                        !vch.room.isEqual(victim.room)
                    ) {
                        break;
                    }

                    if (victim.room.flags & Golem.RoomFlags.ROOM_SAFE) {
                        break;
                    }

                    found = true;

                    let damage = ~~(Math.random() * 2);
                    let damageType = Golem.Combat.DamageTypeBash;

                    let weapon = vch.getEquipment(Golem.WearLocations.WearLocationWielded);
                    if(!weapon) {
                        damage += ~~(Math.random() * (vch.strength / 3));

                        const unarmedCombatProficiency =
                            vch.findProficiencyByName('unarmed combat');
                        /* TODO: check if wielding or not! ... weapon type profs.. */
                        if (unarmedCombatProficiency) {
                            /* +1 damage to unarmed base damage for every 10% of unarmed combat proficiency */
                            damage += Math.floor(
                                unarmedCombatProficiency.proficiency / 10
                            );
                        }
                    } else {
                        let sum = 0;

                        const v0 = parseInt(weapon.value0),
                            v1 = parseInt(weapon.value1),
                            v2 = parseInt(weapon.value2),
                            v3 = parseInt(weapon.value3);

                        for(let i = 0; i < v0; i++) {
                            sum += ~~(Math.random() * v1);
                        }

                        sum += v2;

                        damageType = v3;
                        damage = sum;
                    }

                    /* Check victim dodge skill */
                    const victimDodgeProficiency =
                        victim.findProficiencyByName('dodge');
                    if (victimDodgeProficiency) {
                        if (
                            Math.random() <
                            victimDodgeProficiency.proficiency / 100 / 5
                        ) {
                            vch.send('{D' +
                                victim.getShortDescriptionUpper(vch) +
                                    '{D dodges out of the way of your attack!{x\r\n'
                            );
                            victim.send(
                                '{DYou dodge an attack by ' +
                                    vch.getShortDescription(victim) +
                                    '{D!{x\r\n'
                            );
                            continue;
                        }
                    }

                    const armorClass = victim.getArmorValues();

                    this.damage(
                        vch,
                        victim,
                        true,
                        Math.max(0, damage - (armorClass[damageType] / 4)),
                        damageType
                    );

                    if(victim.group) {
                        for(let iter = victim.group.head; iter != null; iter = iter.next) {
                            const gch = iter.value;

                            if(!gch.fighting && vch.room && gch.room && vch.room.isEqual(gch.room)) {
                                gch.send('{WYou start attacking ' + vch.getShortDescription(gch) + '{W in defense of ' + victim.getShortDescription(gch) + "{W!{x\r\n");
                                gch.fighting = vch;
                                gch.combat = victim.combat;

                                if(gch.combat && gch.combat.participants) {
                                    gch.combat.participants.push(gch);
                                }
                            }
                        }
                    } else {
                        if(!victim.fighting || !victim.combat) {
                            victim.fighting = vch;
                            victim.combat = vch.combat;
                        }
                    }
                }
            }
        }

        if (!found) {
            this.disposeCombat(combat);
        }
    }
}

Golem.registerEventHandler('combatUpdate', onCombatUpdate);
