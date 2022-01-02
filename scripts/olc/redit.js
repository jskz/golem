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
{Gredit flag <flag name> - {gToggle a room flag
{Gredit save             - {gSave room to database{x
`);
    }

    if (!ch.room) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, rest] = Golem.util.oneArgument(args);

    switch (firstArgument) {
        // update the room name with remaining argument string, if there is any
        case 'name':
            if (!rest.length) {
                ch.send("A room name argument is required.\r\nExample: redit name New Room Name Here\r\n");
                return;
            }

            ch.room.name = rest;
            ch.send("Ok.\r\n");
            return;

        case 'flag':
            if (!rest.length) {
                ch.send("A room flag argument to toggle is required.\r\nExample: redit flag safe\n");
                return;
            }

            const flag = Golem.util.findRoomFlag(rest);
            if(!flag) {
                ch.send("No such room flag exists.\r\n");
                return;
            }

            if(!(ch.room.flags & flag.flag)) {
                ch.room.flags |= flag.flag;
                ch.send("Ok.  Enabled room flag " + flag.name + "\r\n");
                return;
            }

            ch.room.flags &= ~(flag.flag);
            ch.send("Ok.  Disabled room flag " + flag.name + ".\r\n");
            return;

        // use the string editor utility to edit the room description string, updating it on completion
        case 'description':
            Golem.StringEditor(ch.client,
                ch.room.description,
                (_, string) => {
                    ch.room.description = string;
                    ch.send("Ok.\r\n");
                });
            return;

        // try to save room to database
        case 'save':
            const err = ch.room.save();
            if (!err) {
                ch.send("Ok.\r\n");
                return;
            }

            ch.send("Something went wrong trying to save this room: " + v + "\r\n");
            return;

        default:
            displayUsage();
            return;
    }
}

Golem.registerPlayerCommand('redit', do_redit, Golem.Levels.LevelBuilder);