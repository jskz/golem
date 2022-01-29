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
{Gxedit delete <direction>           - {gBi-directionally delete an exit
{Gxedit dig <direction>              - {gTry to create a dig a new room
{Gxedit flag <direction> <flag name> - {gToggle a flag for a given direction
{Gxedit link <direction> <id>        - {gCreate a two-way exit to an existing room
{Gxedit unlink <direction>           - {gUnlink this room's side of an exit
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
        case 'delete':
            {
                let [direction, _] = Golem.util.oneArgument(rest);

                if(!VALID_DIRECTIONS.includes(direction)) {
                    ch.send("That's not a valid direction.\r\n");
                    return;
                }

                const dir = DIRECTION_TO_VALUE[direction];
                if(!ch.room.exit[dir]) {
                    ch.send("There is no exit in that direction here.\r\n");
                    return;
                }

                const reverseExit = ch.room.exit[dir].to.exit[Golem.util.reverseDirection[dir]];
                if(reverseExit) {
                    if(reverseExit.delete()) {
                        ch.send("Failed to delete reverse exit.\r\n");
                        return;
                    }

                    delete ch.room.exit[dir].to.exit[Golem.util.reverseDirection[dir]];
                }

                if(ch.room.exit[dir].delete()) {
                    ch.send("Failed to delete exit.\r\n");
                    return;
                }

                delete ch.room.exit[dir];
                ch.send("Ok.\r\n");
                return;
            }

        case 'unlink':
            {
                let [direction, _] = Golem.util.oneArgument(rest);

                if(!VALID_DIRECTIONS.includes(direction)) {
                    ch.send("That's not a valid direction.\r\n");
                    return;
                }

                const dir = DIRECTION_TO_VALUE[direction];
                if(!ch.room.exit[dir]) {
                    ch.send("There is no exit in that direction here.\r\n");
                    return;
                }

                if(ch.room.exit[dir].delete()) {
                    ch.send("Failed to delete exit.\r\n");
                    return;
                }

                delete ch.room.exit[dir];
                ch.send("Ok.\r\n");
                return;
            }

        case 'dig':
            {
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
                return;
            }

        case 'link':
            {
                let [direction, xss] = Golem.util.oneArgument(rest);

                if(!VALID_DIRECTIONS.includes(direction)) {
                    ch.send("That's not a valid direction.\r\n");
                    return;
                }

                const dir = DIRECTION_TO_VALUE[direction];
                if(ch.room.exit[dir]) {
                    ch.send("You can't link with a direction that already has an exit here.\r\n");
                    return;
                }

                const toRoom = Golem.game.loadRoomIndex(parseInt(xss));
                if(!toRoom) {
                    ch.send("Failed to find that existing room index.\r\n");
                    return;
                }

                if(toRoom.exit[Golem.util.reverseDirection[dir]]) {
                    ch.send("That room has an existing exit in the reverse-direction from this direction.\r\n");
                    return;
                }

                const exit = Golem.NewExit(
                    ch.room,
                    dir,
                    toRoom,
                    0
                );

                if(exit.finalize()) {
                    ch.send("Something went wrong trying to save an exit to a new room.\r\n");
                    return;
                }

                ch.room.exit[dir] = exit;
                const reverseExit = Golem.NewExit(
                    toRoom,
                    Golem.util.reverseDirection[dir],
                    ch.room,
                    0
                );

                if(reverseExit.finalize()) {
                    ch.send("Something went wrong trying to save a reverse-exit from the new room to your current room.\r\n");
                    return;
                }

                toRoom.exit[Golem.util.reverseDirection[dir]] = reverseExit;
                ch.send("Ok.\r\n");
                return;
            }

        case 'flag':
            {
                let [direction, flagName] = Golem.util.oneArgument(rest);

                if(!VALID_DIRECTIONS.includes(direction)) {
                    ch.send("That's not a valid direction.\r\n");
                    return;
                }

                const dir = DIRECTION_TO_VALUE[direction];
                const exitFlag = Golem.util.findExitFlag(flagName);

                if(!exitFlag) {
                    ch.send("That's not a valid exit flag.\r\n");
                    return;
                }

                if(ch.room.exit[dir].flags & exitFlag.flag) {
                    ch.room.exit[dir].flags &= ~exitFlag.flag;

                    if(ch.room.exit[dir].save()) {
                        ch.send("Something went wrong trying to save this exit flag update.\r\n");
                        return;
                    }

                    ch.send("Ok.  Disabled " + exitFlag.name + " on exit " + direction + ".\r\n");
                    return;
                }

                ch.room.exit[dir].flags |= exitFlag.flag;
                if(ch.room.exit[dir].save()) {
                    ch.send("Something went wrong trying to save this exit flag update.\r\n");
                    return;
                }

                ch.send("Ok.  Enabled " + exitFlag.name + " on exit " + direction + ".\r\n");
                return;
            }

        default:
            displayUsage();
            break;
    }
}

Golem.registerPlayerCommand('xedit', do_xedit, Golem.Levels.LevelBuilder);