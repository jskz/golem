/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_zedit(ch, args) {
    function displayUsage() {
        ch.send(
            `{WZone editor usage:

{Gzones               - {gDisplay a list of all zones
{Gzedit save          - {gSave zone properties to database
{Gzedit create        - {gCreate a new zone

{WThe following values may be used in a general way with the syntax:
{Gzedit <zone_id> <attribute> <value>
{Wto set a zone's specified attribute to the given value.

{GThe following attributes are available:{g
    name who_description reset_message reset_frequency
{x`);
    }

    let [firstArgument, xs] = Golem.util.oneArgument(args);
    let [secondArgument, xxs] = Golem.util.oneArgument(xs);

    if(!args.length) {
        displayUsage();
        return;
    }

    switch (firstArgument) {
        case 'create':
            const newZone = Golem.game.createZone();
            if(!newZone) {
                ch.send("Something went wrong trying to create a new zone.\r\n");
                return;
            }

            ch.send(`Created a new zone with ID ${newZone.id}.\r\n`);
            return;

        case 'save':
            if (!ch.room
            || ch.room.flags & Golem.RoomFlags.ROOM_PLANAR
            || ch.room.flags & Golem.RoomFlags.ROOM_VIRTUAL) {
                ch.send("You can't do that here.\r\n");
                return;
            }

            if(ch.room.zone.save()) {
                ch.send("Something went wrong trying to save this zone.\r\n");
                return;
            }

            ch.send("Ok.\r\n");
            return;

        default:
            {
                const zoneId = parseInt(firstArgument);
                const zone = Golem.game.findZoneByID(zoneId);

                if(!zone) {
                    ch.send("No such zone.\r\n");
                    return;
                }

                switch(secondArgument) {
                    case 'name':
                        zone.name = xxs;
                        ch.send("Ok.\r\n");
                        break;

                    case 'who_description':
                        zone.whoDescription = xxs;
                        ch.send("Ok.\r\n");
                        break;

                    case 'reset_frequency':
                        const newFrequency = parseInt(xxs);

                        if(isNaN(newFrequency)) {
                            ch.send("Please provide an integer reset frequency in minutes.\r\n");
                            return;
                        }

                        zone.resetFrequency = newFrequency;
                        ch.send("Ok.\r\n");
                        break;

                    case 'reset_message':
                        zone.resetMessage = xxs;
                        ch.send("Ok.\r\n");
                        break;

                    default:
                        displayUsage();
                        return;
                }
            }
    }
}

Golem.registerPlayerCommand('zedit', do_zedit, Golem.Levels.LevelAdmin);