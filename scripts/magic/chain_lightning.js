/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function spell_chain_lightning(ch, args) {
    function performChainLightningAttack(target, maxRecursion = 3) {
        if(!maxRecursion) {
            return;
        }
        
        for (let iter = target.room.characters.head; iter !== null; iter = iter.next) {
            const rch = iter.value;
    
            if (!rch.isEqual(target)) {
                rch.send(
                    '{Y' +
                        target.getShortDescriptionUpper(rch) +
                        '{Y is struck by a crackling arc of chain lightning!{x\r\n'
                );
            }
        }

        target.send('{YYou are struck by a crackling arc of chain lightning!{x\r\n');

        const amount = ~~(Math.random() * 50) + 10;
        Golem.game.damage(ch, target, false, amount, Golem.Combat.DamageTypeExotic);
        if(!target || !target.room.isEqual(ch.room)) {
            return;
        }

        // 40% chance to recurse to a random other group member of the target, if the target is in a group
        if(Math.random() >= 0.6 && target.group) {
            const otherGroupMembers = target.group.values().filter(groupMember => !groupMember.isEqual(target) && groupMember.room?.isEqual(ch.room));
            
            // only if there are other members in the party - can't chain to yourself, but it is valid to
            // chain back & forth until recursive depth is exhausted
            if(otherGroupMembers.length) {
                const choice = ~~(Math.random() * otherGroupMembers.length);

                for (let iter = target.room.characters.head; iter !== null; iter = iter.next) {
                    const rch = iter.value;
            
                    if (!rch.isEqual(target)) {
                        rch.send(
                            '{WThe magical bolt arcs from ' +
                                target.getShortDescription(rch) +
                                '{W to ' + otherGroupMembers[choice].getShortDescription(rch) + '{W with an instantaneous, ear-splitting BANG!{x\r\n'
                        );
                    }
                }
                
                return performChainLightningAttack(otherGroupMembers[choice], maxRecursion - 1);
            }
        }
    }

    if (!args.length) {
        ch.send('This spell requires a target.\r\n');
        return;
    }

    const target = ch.fighting || ch.findCharacterInRoom(args);
    if (!target || !ch.room || !target.room || target.isEqual(ch) || !target.room.isEqual(ch.room)) {
        ch.send("Your target couldn't be found.\r\n");
        return;
    }
    
    // it's possible for "chain lightning" to travel up to 7 times, if the caster has this spell mastered
    performChainLightningAttack(target, ~~(this.proficiency / 20) + 2);
}

Golem.registerSpellHandler('chain lightning', spell_chain_lightning);
