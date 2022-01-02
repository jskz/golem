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

{Greset list - {gDisplay all resets for the current room
{Greset room - {gTrigger a reset for this room immediately
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
        case 'list':
            let i = 1;
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