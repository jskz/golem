/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
function do_xedit(ch, args) {
    const VALID_DIRECTIONS = ['north', 'east', 'south', 'west', 'up', 'down'];
    const DIRECTION_TO_VALUE = {
        'north': Golem.Directions.DirectionNorth,
        'east': Golem.Directions.DirectionEast,
        'south': Golem.Directions.DirectionSouth,
        'west': Golem.Directions.DirectionWest,
        'up': Golem.Directions.DirectionUp,
        'down': Golem.Directions.DirectionDown
    };

    function displayUsage() {
        ch.send(
            `{WExit editor usage:
{Gxedit delete <direction> - {gDestroy a permanent exit
{Gxedit dig <direction> - {gTry to create a new zone room in a direction
{Gxedit flag <direction> <flag name> - {gToggle a flag for a given direction
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

            const dir = DIRECTION_TO_VALUE[direction];
            if(ch.room.exit[dir]) {
                ch.send("You can't dig in a direction that already has an exit.\r\n");
                return;
            }

            const newRoom = ch.room.zone.createRoom();

            if(!newRoom) {
                ch.send("Something went wrong trying to dig a new room.\r\n");
                return;
            }

            const exit = Golem.NewExit(
                ch.room,
                dir,
                newRoom,
                0
            );

            if(exit.finalize()) {
                ch.send("Something went wrong trying to save an exit to a new room.\r\n");
                return;
            }

            ch.room.exit[dir] = exit;
            const reverseExit = Golem.NewExit(
                newRoom,
                Golem.util.reverseDirection[dir],
                ch.room,
                0
            );

            if(reverseExit.finalize()) {
                ch.send("Something went wrong trying to save a reverse-exit from the new room to your current room.\r\n");
                return;
            }

            newRoom.exit[Golem.util.reverseDirection[dir]] = reverseExit;
            ch.send("Ok.\r\n");
            break;

        default:
            displayUsage();
            break;
    }
}

Golem.registerPlayerCommand('xedit', do_xedit, Golem.Levels.LevelBuilder);