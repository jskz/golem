/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function onCombatUpdate() {
    for(let iter = this.fights.head; iter != null; iter = iter.next) {
        const combat = iter.value;
        
        let found = false;

        for(let i = 0; i < combat.participants.length; i++) {
            const vch = combat.participants[i];

            if(vch.room === null
            || vch.fighting.room === null
            || vch.room.id != vch.fighting.room.id) {
                continue;
            }

            found = true;

            let attackerRounds = 1,
                dexterityBonusRounds = parseInt((vch.dexterity - 10) / 4);

            attackerRounds += dexterityBonusRounds;

            for(let r = 0; r < attackerRounds; r++) {
                let victim = vch.fighting;

                if(!victim) {
                    break;
                }

                let damage = ~~(Math.random() * 2);

                damage += ~~(Math.random() * (vch.strength / 3));

                const unarmedCombatProficiency = vch.findProficiencyByName('unarmed combat');
                /* TODO: check if wielding or not! ... weapon type profs.. */
                if(unarmedCombatProficiency) {
                    /* +1 damage to unarmed base damage for every 10% of unarmed combat proficiency */
                    damage += Math.floor(unarmedCombatProficiency.proficiency / 10);
                }

                /* Check victim dodge skill */
                const victimDodgeProficiency = victim.findProficiencyByName('dodge');
                if(victimDodgeProficiency) {
                    vch.send("Victim has dodge proficiency ...\r\n" + unarmedCombatProficiency.proficiency);
                    victim.send("You have dodge proficiency ...\r\n" + unarmedCombatProficiency.proficiency);
                    if(Math.random() < (unarmedCombatProficiency.proficiency / 100) / 5) {
                        vch.send(victim.getShortDescriptionUpper(vch) + " dodges out of the way of your attack!\r\n");
                        victim.send("You dodge an attack by " + vch.getShortDescription(victim) + "!\r\n");
                        continue;
                    }
                }

                this.damage(vch, victim, true, damage, Golem.Combat.DamageTypeBash);
            }
        }

        if(!found) {
            this.disposeCombat(combat);
            break;
        }
    }
}

Golem.registerEventHandler('combatUpdate', onCombatUpdate);