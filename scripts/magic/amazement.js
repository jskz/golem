/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */

// It's like a basement, but a maze!
function spell_amazement(ch) {
    if(!ch.room) {
        ch.send("You can't do that here.\r\n");
        return;
    }

    if(ch.room.exit[Golem.Directions.DirectionDown]) {
        ch.send("You can't burrow a maze into a room above another.\r\n");
        return;
    }

    const maze = Golem.game.generateDungeon(2, 32, 32);
    if(ch.room.flags & Golem.RoomFlags.ROOM_PLANAR) {
        ch.room.plane.map.layers[ch.room.z].terrain[ch.room.y][ch.room.x] = Golem.TerrainTypes.TerrainTypeCaveDeepWall1;
    }

    ch.send("{YA crack of lightning sunders the earth before you, revealing a dungeon!{x\r\n");

    try {
        const mazeFirstFloor = maze.floors[0];
        const mazeEntrance =
            mazeFirstFloor.grid[mazeFirstFloor.entryX][
                mazeFirstFloor.entryY
            ].room;

        ch.room.exit[Golem.Directions.DirectionDown] = Golem.NewExit(
            Golem.Directions.DirectionDown,
            mazeEntrance,
            Golem.ExitFlags.EXIT_IS_DOOR | Golem.ExitFlags.EXIT_CLOSED
        );

        mazeEntrance.exit[Golem.Directions.DirectionUp] =
            Golem.NewExit(
                Golem.Directions.DirectionUp,
                ch.room,
                Golem.ExitFlags.EXIT_IS_DOOR |
                    Golem.ExitFlags.EXIT_CLOSED
            );
    } catch(err) {
        ch.send(err.toString());
    }
}

Golem.registerSpellHandler('amazement', spell_amazement);
