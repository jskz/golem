/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_mstat(ch, args) {
    function resource(current, maximum) {
        const percentage = Golem.util.resourcePercentage(current, maximum);

        return `${Golem.util.severityColourFromPercentage(percentage)}${current}{c/{C${maximum}`;
    }

    function stat(target, statType) {
        const value = Golem.util.characterStat(target, statType);

        return `{C${value.modified}{c/{C${value.base}{c`;
    }

    function appendEffects(output, target) {
        if (!target.effects || !target.effects.count) {
            return;
        }

        for (let iter = target.effects.head; iter !== null; iter = iter.next) {
            output.push(`{CEffect:{c {C${Golem.util.effectDescription(iter.value)}`);
        }
    }

    function formatOutput(target) {
        const raceName = target.race && target.race.displayName ? target.race.displayName : 'unknown race';
        const jobName = target.job && target.job.displayName ? target.job.displayName : 'unknown job';
        const inventoryCount = target.inventory ? target.inventory.count : 0;
        const effectCount = target.effects ? target.effects.count : 0;
        const output = [];

        output.push(
            `{cCharacter {C'${Golem.util.characterName(target)}'{c is type {C${Golem.util.characterTypeName(target)}{c, ` +
                `level {C${target.level}{c {C${raceName}{c {C${jobName}{c.{x`
        );
        output.push(
            `{CId:{c ${target.id}  {CPosition:{c ${Golem.util.positionName(target.position)}  ` +
                `{CLocation:{c ${Golem.util.characterLocationName(target)}`
        );
        output.push(`{CFlags:{c ${Golem.util.characterFlagNames(target.flags)}`);
        output.push(`{CAffected:{c ${Golem.util.affectedFlagNames(target.affected)}`);
        output.push(
            `{CResources:{c health ${resource(target.health, target.maxHealth)}  ` +
                `{cmana ${resource(target.mana, target.maxMana)}  ` +
                `{cstamina ${resource(target.stamina, target.maxStamina)}`
        );
        output.push(
            `{CAttributes (mod/base):{c str ${stat(target, Golem.StatTypes.STAT_STRENGTH)}  ` +
                `DEX ${stat(target, Golem.StatTypes.STAT_DEXTERITY)}  ` +
                `CON ${stat(target, Golem.StatTypes.STAT_CONSTITUTION)}  ` +
                `INT ${stat(target, Golem.StatTypes.STAT_INTELLIGENCE)}  ` +
                `WIS ${stat(target, Golem.StatTypes.STAT_WISDOM)}  ` +
                `CHA ${stat(target, Golem.StatTypes.STAT_CHARISMA)}  ` +
                `LCK ${stat(target, Golem.StatTypes.STAT_LUCK)}`
        );
        output.push(
            `{CDefense:{c ${target.defense}  {CGold:{c ${target.gold}  ` +
                `{CExperience:{c ${target.experience}  {CPractices:{c ${target.practices}`
        );
        output.push(
            `{CFighting:{c ${Golem.util.characterName(target.fighting)}  ` +
                `{CFollowing:{c ${Golem.util.characterName(target.following)}  ` +
                `{CLeader:{c ${Golem.util.characterName(target.leader)}`
        );
        output.push(`{CInventory:{c ${inventoryCount} carried  {CEffects:{c ${effectCount} active`);

        if (!(target.flags & Golem.CharacterFlags.CHAR_IS_PLAYER)) {
            output.push(`{CShort:{c ${target.shortDescription || ''}`);
            output.push(`{CLong:{c ${target.longDescription || ''}`);
        }

        appendEffects(output, target);

        return output.join('\r\n') + '{x\r\n';
    }

    const argument = args.trim();
    if (!argument.length) {
        ch.send('Mstat whom?\r\n');
        return;
    }

    const target = Golem.util.findCharacter(ch, argument);
    if (!target) {
        ch.send('No such character.\r\n');
        return;
    }

    ch.send(formatOutput(target));
}

Golem.registerPlayerCommand('mstat', do_mstat, Golem.Levels.LevelHero + 1);
