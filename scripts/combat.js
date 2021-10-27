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

        for (let i = 0; i < combat.participants.length; i++) {
            const vch = combat.participants[i];

            if (vch.Room === null) {
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
                    vch.Room.id != victim.room.id
                ) {
                    break;
                }

                if (victim.room.flags & Golem.RoomFlags.ROOM_SAFE) {
                    break;
                }

                try {
                    found = true;

                    let damage = ~~(Math.random() * 2);

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

                    this.damage(
                        vch,
                        victim,
                        true,
                        damage,
                        Golem.Combat.DamageTypeBash
                    );
                } catch (err) {
                    Golem.game.broadcast(err.toString());
                }

                /*
                if(victim && victim.group !== null) {
                    for(let iter = victim.gch.head; iter != null; iter = gch.next) {
                        const gch = iter.value;

                        if(!gch.fighting) {
                            gch.send('{WYou start attacking ' + ch.getShortDescription(gch) + '{W in defense of ' + victim.getShortDescriptionUpper(gch) + '{W!{x');
                            gch.fighting = ch;
                            gch.combat = ch.combat;
                            gch.combat.insert(gch);
                        }
                    }
                }
                */
            }
        }

        if (!found) {
            this.disposeCombat(combat);
            break;
        }
    }
}

Golem.registerEventHandler('combatUpdate', onCombatUpdate);
