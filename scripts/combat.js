function onCombatUpdate() {
    for(let iter = this.fights.head; iter != null; iter = iter.next) {
        const combat = iter.value;
        
        let found = false;

        for(let i = 0; i < combat.participants.length; i++) {
            const vch = combat.participants[i];

            if(vch.room === null 
            || vch.fighting === null
            || vch.fighting.room === null
            || vch.room.id != vch.fighting.room.id) {
                continue;
            }

            found = true;

            let attackerRounds = 1,
                dexterityBonusRounds = parseInt((vch.dexterity - 10) / 4);

            attackerRounds += dexterityBonusRounds;

            for(let r = 0; r < attackerRounds; r++) {
                let damage = ~~(Math.random() * 2);
                damage += ~~(Math.random() * (vch.strength / 3));

                this.damage(vch, vch.fighting, true, damage, Golem.Combat.DamageTypeBash);
            }
        }

        if(!found) {
            this.disposeCombat(combat);
            break;
        }
    }
}

Golem.registerEventHandler('combatUpdate', onCombatUpdate);