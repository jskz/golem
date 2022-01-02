/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_redit(ch, args) {
    function displayUsage() {
        ch.send(
            `Usage:
    redit name <room name> - Set room name
    redit description - String editor for room description
    redit save - Save room to database
`);
    }

    if (!ch.room) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, rest] = Golem.util.oneArgument(args);

    switch (firstArgument) {
        case 'name':
            if (!rest.length) {
                ch.send("A room name argument is required.\r\nExample: redit name New Room Name Here\r\n");
                break;
            }

            ch.room.name = rest;
            ch.send("Ok.\r\n");
            break;

        case 'save':
            const err = ch.room.save();
            if (!err) {
                ch.send("Ok.\r\n");
                break;
            }

            ch.send("Something went wrong trying to save this room: " + v + "\r\n");
            break;

        default:
            displayUsage();
            break;
    }
}

Golem.registerPlayerCommand('redit', do_redit);