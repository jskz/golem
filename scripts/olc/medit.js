/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_medit(ch, args) {
    function displayUsage() {
        ch.send(
            `{WMobile editor usage:
                
{Gmedit <target> save        - {gSave mobile instance properties globally
{Gmedit <target> description - {gString editor for mobile's description

{WThe following values may be used in a general way with the syntax:
{Gmedit <target> <attribute> <value>
{Wto set that target's specified attribute to the given value.

{GThe following attributes are available:{g
  name health max_health mana max_mana stamina max_stamina gold level
  strength dexterity intelligence wisdom constitution charisma luck
  short_description long_description
{x`);
    }

    if (!ch.room) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, xs] = Golem.util.oneArgument(args);
    let [secondArgument, xxs] = Golem.util.oneArgument(xs);

    if(!args.length) {
        displayUsage();
        return;
    }

    const target = ch.findCharacterInRoom(firstArgument);
    if(!target) {
        ch.send("No such character here.\r\n");
        return;
    }

    switch (secondArgument) {
        case 'save':
            if(target.flags & Golem.CharacterFlags.CHAR_IS_PLAYER) {
                ch.send("Failed: not an NPC.\r\n");
                return;
            }

            if(target.sync()) {
                ch.send("Failed.\r\n");
                return;
            }

            ch.send("Ok.\r\n");
            return;

        case 'name':
            if(target.flags & Golem.CharacterFlags.CHAR_IS_PLAYER) {
                ch.send("Failed: not an NPC.\r\n");
                return;
            }

            target.name = xxs;
            ch.send("Ok.\r\n");
            break;

        case 'short_description':
            if(target.flags & Golem.CharacterFlags.CHAR_IS_PLAYER) {
                ch.send("Failed: not an NPC.\r\n");
                return;
            }

            target.shortDescription = xxs;
            ch.send("Ok.\r\n");
            break;

        case 'long_description':
            if(target.flags & Golem.CharacterFlags.CHAR_IS_PLAYER) {
                ch.send("Failed: not an NPC.\r\n");
                return;
            }

            target.longDescription = xxs;
            ch.send("Ok.\r\n");
            break;

        case 'description':
            Golem.StringEditor(ch.client,
                target.description,
                (_, string) => {
                    target.description = string;
                    ch.send("Ok.\r\n");
                });    
            break;

        case 'level':
            if(target.flags & Golem.CharacterFlags.CHAR_IS_PLAYER) {
                ch.send("Failed: not an NPC.\r\n");
                return;
            }

            const level = parseInt(xxs);
            if(level < 1 || level > Golem.Levels.LevelHero) {
                ch.send("Invalid NPC level provided.\r\n");
                return;
            }

            target.level = level;
            ch.send("Ok.\r\n");
            break;
            
        case 'health':
            target.health = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'max_health':
            target.maxHealth = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'mana':
            target.mana = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'max_mana':
            target.maxMana = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;
                
        case 'stamina':
            target.stamina = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'max_stamina':
            target.maxStamina = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'gold':
            target.gold = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;
    
        case 'strength':
            target.stats[Golem.StatTypes.STAT_STRENGTH] = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'dexterity':
            target.stats[Golem.StatTypes.STAT_DEXTERITY] = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'intelligence':
            target.stats[Golem.StatTypes.STAT_INTELLIGENCE] = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'wisdom':
            target.stats[Golem.StatTypes.STAT_WISDOM] = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'constitution':
            target.stats[Golem.StatTypes.STAT_CONSTITUTION] = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'charisma':
            target.stats[Golem.StatTypes.STAT_CHARISMA] = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        case 'luck':
            target.stats[Golem.StatTypes.STAT_LUCK] = parseInt(xxs);
            ch.send("Ok.\r\n");
            break;

        default:
            displayUsage();
            return;
    }
}

Golem.registerPlayerCommand('medit', do_medit, Golem.Levels.LevelBuilder);