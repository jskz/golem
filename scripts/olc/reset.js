/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */

function do_reset(ch, args) {
    function displayUsage() {
        ch.send(
            `{WReset editor usage:

{Greset list                         - {gDisplay all resets for this room
{Greset room                         - {gTrigger a reset for this room now
{Greset delete <#>                   - {gDelete a reset by its {Greset list{g position{x
{Greset mob <id> <unused> <room max> - {gCreate and save a new NPC reset here{x
{Greset obj <id> <unused> <room max> - {gCreate and save a new object reset here{x
`);
    }
    
    if (!ch.room
    || ch.room.flags & Golem.RoomFlags.ROOM_PLANAR
    || ch.room.flags & Golem.RoomFlags.ROOM_VIRTUAL) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    let [firstArgument, rest] = Golem.util.oneArgument(args);

    function getRoomResetByCount(count) {
        let i = 1;

        for(let iter = ch.room.resets.head; iter != null; iter = iter.next) {
            const reset = iter.value;

            if(i === count) {
                return reset;
            }

            i++;
        }

        return null;
    }

    switch (firstArgument) {
        case 'obj':
        case 'mob':
            {
                const [entityId, xs] = Golem.util.oneArgument(rest);
                const [_, xxs] = Golem.util.oneArgument(xs);
                const [maxInRoom] = Golem.util.oneArgument(xxs);

                const result = ch.room.createReset(firstArgument === 'mob' ? 0 : 1, entityId, -1, maxInRoom);
                if(result) {
                    ch.room.resets.insert(result);
                    ch.send("Ok.\r\n");
                    return;
                }

                ch.send("Failed.\r\n");
                return;
            }

        case 'delete':
            {
                if(!ch.room.resets.count) {
                    ch.send("There are no resets in this room.\r\n");
                    return;
                }

                const reset = getRoomResetByCount(parseInt(rest));

                if(!reset) {
                    ch.send("Failed: could not find a room reset with that list index.\r\n");
                    return;
                }

                if(reset.delete()) {
                    ch.send("Failed.\r\n");
                    return;
                }

                ch.room.resets.remove(reset);
                ch.send("Ok.\r\n");
                return;
            }

        case 'list':
            let i = 1;

            if(!ch.room.resets.count) {
                ch.send("There are no resets in this room.\r\n");
                return;
            }

            for(let iter = ch.room.resets.head; iter != null; iter = iter.next) {
                const reset = iter.value;

                switch(reset.resetType) {
                    case 0:
                        ch.send("#" + i + ". mobile reset with ID " + reset.value0 + ", not exceeding a count of " + reset.value2 + "\r\n");
                        break;
                    case 1:
                        ch.send("#" + i + ". object reset with ID " + reset.value0 + ", not exceeding a count of " + reset.value2 + "\r\n");
                        break;
                    default:
                        break;
                }

                i++;
            }
            break;

        case 'room':
            Golem.game.resetRoom(ch.room);
            ch.send("You reset the room.\r\n");
            break;

        default:
            displayUsage();
            break;
    }
}

Golem.registerPlayerCommand('reset', do_reset, Golem.Levels.LevelBuilder);