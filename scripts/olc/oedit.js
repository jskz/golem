/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_oedit(ch, args) {
    function displayUsage() {
        ch.send(
            `{WObject editor usage:
                
{Goedit <obj> save             - {gSave object instance properties globally
{Goedit <obj> description      - {gString editor for object's description
{Goedit <obj> flag <flag name> - {gToggle object flag by name

{WThe following values may be used in a general way with the syntax:
{Goedit <target> <attribute> <value>
{Wto set the target object's specified attribute to the given value.

{GThe following attributes are available:{g
  name short_description long_description item_type
  value0 value1 value2 value3
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

    try {
        let target = ch.findObjectOnSelf(firstArgument);
        if(!target) {
            target = ch.findObjectInRoom(firstArgument);

            if(!target) {
                ch.send("No such object here.\r\n");
                return;
            }
        }

        switch (secondArgument) {
            case 'save':
                if(target.sync()) {
                    ch.send("Failed.\r\n");
                    return;
                }

                ch.send("Ok.\r\n");
                return;

            case 'item_type':
                // validate item type is valid...

                target.itemType = xxs;
                ch.send("Ok.\r\n");
                break;

            case 'flag':
                if (!xxs.length) {
                    ch.send("An object flag argument to toggle is required.\r\nExample: oedit amulet flag glow\n");
                    return;
                }
    
                const flag = Golem.util.findObjectFlag(xxs);
                if(!flag) {
                    ch.send("No such object flag exists.\r\n");
                    return;
                }
    
                if(!(target.flags & flag.flag)) {
                    target.flags |= flag.flag;
                    ch.send("Ok.  Enabled object flag " + flag.name + " on " + target.getShortDescription(ch) + ".\r\n");
                    return;
                }
    
                target.flags &= ~(flag.flag);
                ch.send("Ok.  Disabled object flag " + flag.name + " on " + target.getShortDescription(ch) + ".\r\n");
                return;
        
            case 'name':
                target.name = xxs;
                ch.send("Ok.\r\n");
                break;

            case 'short_description':
                target.shortDescription = xxs;
                ch.send("Ok.\r\n");
                break;

            case 'long_description':
                target.longDescription = xxs;
                ch.send("Ok.\r\n");
                break;
                
            case 'value0':
                target.value0 = parseInt(xxs);
                ch.send("Ok.\r\n");
                break;

            case 'value1':
                target.value1 = parseInt(xxs);
                ch.send("Ok.\r\n");
                break;

            case 'value2':
                target.value2 = parseInt(xxs);
                ch.send("Ok.\r\n");
                break;

            case 'value3':
                target.value3 = parseInt(xxs);
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

            default:
                displayUsage();
                return;
        }
    } catch(err) {
        ch.send(err.toString());
    }
}

Golem.registerPlayerCommand('oedit', do_oedit, Golem.Levels.LevelBuilder);