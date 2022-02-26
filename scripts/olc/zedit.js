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

{Gzones                 - {gDisplay a list of all zones
{Gzedit <zone_id> save  - {gSave a zone's properties to database
{Gzedit create          - {gCreate a new zone

{WThe following values may be used in a general way with the syntax:
{Gzedit <zone_id> <attribute> <value>
{Wto set a zone's specified attribute to the given value.

{GThe following attributes are available:{g
    name who_description reset_message reset_frequency lo_hi
{x`);
    }

    let [firstArgument, xs] = Golem.util.oneArgument(args);
    let [secondArgument, xxs] = Golem.util.oneArgument(xs);

    if (!args.length) {
        displayUsage();
        return;
    }

    switch (firstArgument) {
        case 'create':
            const newZone = Golem.game.createZone();
            if (!newZone) {
                ch.send("Something went wrong trying to create a new zone.\r\n");
                return;
            }

            ch.send(`Created a new zone with ID ${newZone.id}.\r\n`);
            return;

        default:
            {
                const zoneId = parseInt(firstArgument);
                const zone = Golem.game.findZoneByID(zoneId);

                if (!zone) {
                    ch.send("No such zone.\r\n");
                    return;
                }

                switch (secondArgument) {
                    case 'name':
                        zone.name = xxs;
                        ch.send("Ok.\r\n");
                        return;

                    case 'who_description':
                        zone.whoDescription = xxs;
                        ch.send("Ok.\r\n");
                        return;

                    case 'reset_frequency':
                        const newFrequency = parseInt(xxs);

                        if (isNaN(newFrequency)) {
                            ch.send("Please provide an integer reset frequency in minutes.\r\n");
                            return;
                        }

                        zone.resetFrequency = newFrequency;
                        ch.send("Ok.\r\n");
                        return;

                    case 'reset_message':
                        zone.resetMessage = xxs;
                        ch.send("Ok.\r\n");
                        return;

                    case 'lo_hi':
                        let [thirdArgument, xxxs] = Golem.util.oneArgument(xxs);

                        const low = parseInt(thirdArgument);
                        const high = parseInt(xxxs);

                        if (isNaN(low) || isNaN(high)) {
                            ch.send("Please provide two integer values for zone low and high IDs.\r\n");
                            return;
                        }

                        if (!Golem.game.validZoneRange(low, high)) {
                            ch.send("Please provide a valid low-high ID range which does not overlap an existing zone.\r\n");
                            return;
                        }

                        zone.low = low;
                        zone.high = high;

                        ch.send("Ok.\r\n");
                        return;

                    case 'save':
                        if (zone.save()) {
                            ch.send("Something went wrong trying to save this zone.\r\n");
                            return;
                        }

                        ch.send("Ok.\r\n");
                        return;

                    default:
                        displayUsage();
                        return;
                }
            }
    }

    return;
}

Golem.registerPlayerCommand('zedit', do_zedit, Golem.Levels.LevelAdmin);