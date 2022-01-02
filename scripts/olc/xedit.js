/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_xedit(ch, args) {
    const VALID_DIRECTIONS = ['north', 'east', 'south', 'west', 'up', 'down'];

    function displayUsage() {
        ch.send(
            `{WExit editor usage:
{Gxedit delete <direction> - {gDestroy a permanent exit
{Gxedit dig <direction> - {gTry to create a new zone room in a direction
{Gxedit flags <direction> <flag name> - {gToggle a flag for a given direction 
`);
    }

    if (!ch.room
    || ch.room.flags & Golem.RoomFlags.ROOM_PLANAR
    || ch.room.flags & Golem.RoomFlags.ROOM_VIRTUAL) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, rest] = Golem.util.oneArgument(args);

    switch (firstArgument) {
        case 'dig':
            let [direction, _] = Golem.util.oneArgument(rest);

            if(!VALID_DIRECTIONS.includes(direction)) {
                ch.send("That's not a valid direction.\r\n");
                return;
            }

            ch.send("Trying to dig " + direction + "\r\n");
            break;

        default:
            displayUsage();
            break;
    }
}

Golem.registerPlayerCommand('xedit', do_xedit, Golem.Levels.LevelBuilder);