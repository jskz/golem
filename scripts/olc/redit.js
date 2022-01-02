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
            `{WRoom editor usage:
{Gredit name <room name> - {gSet room name
{Gredit description      - {gString editor for room description
{Gredit save             - {gSave room to database{x
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

        case 'description':
            Golem.StringEditor(ch.client,
                ch.room.description,
                (_, string) => {
                    ch.room.description = string;
                    ch.send("Ok.\r\n");
                });
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