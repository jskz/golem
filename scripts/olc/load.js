/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */

// instantiate an object or mobile index
function do_load(ch, args) {
    function displayUsage() {
        ch.send(
            `{WLoad usage:
{Gload mob <mobile id> - {gCreate a mobile instance by id in this room
{Gload obj <object id> - {gCreate an object instance by id on person if ITEM_TAKE else in-room{x
`);
    }
    
    if (!ch.room) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, rest] = Golem.util.oneArgument(args);

    switch (firstArgument) {
        case 'mob':
            {
                if (!rest.length) {
                    ch.send("A mobile ID argument is required.\r\nExample: load mob 1\r\n");
                    return;
                }

                const mob = Golem.game.loadMobileIndex(parseInt(rest));
                if(!mob) {
                    ch.send("Failed.\r\n");
                    return;
                }

                ch.room.addCharacter(mob);
                Golem.game.characters.insert(mob);

                for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
                    const rch = iter.value;

                    if (!rch.isEqual(mob)) {
                        rch.send(
                            '{W' + mob.getShortDescriptionUpper(rch) + '{W appears in a puff of smoke!{x\r\n'
                        );
                    }
                }

                return;
            }
            
        case 'obj':
            {
                if (!rest.length) {
                    ch.send("An object ID argument is required.\r\nExample: load obj 1\r\n");
                    return;
                }

                const obj = Golem.game.newObjectInstance(parseInt(rest));

                if(!obj) {
                    ch.send("Failed.\r\n");
                    return;
                }

                if(obj.flags & Golem.ObjectFlags.ITEM_TAKE) {
                    ch.attachObject(obj);
                    ch.addObject(obj);
                    Golem.game.objects.insert(obj);
                } else {
                    ch.room.addObject(obj);
                    Golem.game.objects.insert(obj);
                }

                for (let iter = ch.room.characters.head; iter !== null; iter = iter.next) {
                    const rch = iter.value;

                    rch.send(
                        '{W' + obj.getShortDescriptionUpper(rch) + '{W appears in a puff of smoke!{x\r\n'
                    );
                }

                return;
            }

        default:
            displayUsage();
            break;
    }
}

Golem.registerPlayerCommand('load', do_load, Golem.Levels.LevelBuilder);