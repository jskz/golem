CREATE TABLE scripts (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE,
    `script` TEXT,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id)
);

CREATE TABLE mobile_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `mobile_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (mobile_id) REFERENCES mobiles(id),
    FOREIGN KEY (script_id) REFERENCES scripts(id)
);

CREATE TABLE object_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `object_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (`object_id`) REFERENCES objects(id),
    FOREIGN KEY (script_id) REFERENCES scripts(id)
);

CREATE TABLE room_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `room_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (room_id) REFERENCES rooms(id),
    FOREIGN KEY (script_id) REFERENCES scripts(id)
);

CREATE TABLE plane_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `plane_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (plane_id) REFERENCES planes(id),
    FOREIGN KEY (script_id) REFERENCES scripts(id)
);

INSERT INTO 
    scripts(id, name, script)
VALUES (1, 'limbo-developer-maze', 
"module.exports = {
    onGenerate: function (plane) {
        // Create a self-referential exit from limbo which is yet to be generated
        const limbo = Golem.game.loadRoomIndex(Golem.KnownLocations.Limbo);

        function populateDungeon(dungeon) {
            for (let z = 0; z < dungeon.floors.length; z++) {
                for (let y = 0; y < dungeon.floors[z].grid.length; y++) {
                    for (let x = 0; x < dungeon.floors[z].grid[y].length; x++) {
                        const cell = dungeon.floors[z].grid[x][y];
                        if (!cell.wall) {
                            const chanceToSpawnCreature = ~~(
                                Math.random() * 100
                            );

                            if (chanceToSpawnCreature > 90) {
                                const baseMobile =
                                    Golem.game.loadMobileIndex(1);

                                try {
                                    if (baseMobile && cell.room) {
                                        baseMobile.name =
                                            'aggressive slime angry agitate agitated';
                                        baseMobile.shortDescription =
                                            'an agitated slime';
                                        baseMobile.longDescription =
                                            'An angry slime has festered to the point of open hostility in the dungeon.';
                                        baseMobile.description =
                                            'This angry slime has one big chip on its gelatinous shoulder.';
                                        baseMobile.dexterity = 15;
                                        baseMobile.health = 100;
                                        baseMobile.maxHealth = 100;
                                        baseMobile.strength = 20;
                                        baseMobile.experience = 6000;
                                        baseMobile.flags =
                                            Golem.CharFlags.CHAR_AGGRESSIVE;

                                        cell.room.addCharacter(baseMobile);
                                        Golem.game.characters.insert(
                                            baseMobile
                                        );
                                    }
                                } catch (err) {
                                    Golem.game.broadcast(err.toString());
                                }
                            }
                        }
                    }
                }
            }
        }

        if (limbo) {
            limbo.exit[Golem.Directions.DirectionDown] = Golem.NewExit(
                Golem.Directions.DirectionDown,
                limbo,
                Golem.ExitFlags.EXIT_IS_DOOR |
                    Golem.ExitFlags.EXIT_CLOSED |
                    Golem.ExitFlags.EXIT_LOCKED
            );

            module.exports.onGenerationComplete = (plane) => {
                const dungeonFirstFloor = plane.dungeon.floors[0];
                const dungeonEntrance =
                    dungeonFirstFloor.grid[dungeonFirstFloor.entryX][
                        dungeonFirstFloor.entryY
                    ].room;

                dungeonEntrance.flags =
                    Golem.RoomFlags.ROOM_VIRTUAL | Golem.RoomFlags.ROOM_SAFE;

                // Tie the dungeon's entrance to limbo and unlock the trapdoor
                limbo.exit[Golem.Directions.DirectionDown].to = dungeonEntrance;
                dungeonEntrance.exit[Golem.Directions.DirectionUp] =
                    Golem.NewExit(
                        Golem.Directions.DirectionUp,
                        limbo,
                        Golem.ExitFlags.EXIT_IS_DOOR |
                            Golem.ExitFlags.EXIT_CLOSED
                    );
                limbo.exit[Golem.Directions.DirectionDown].flags =
                    Golem.ExitFlags.EXIT_IS_DOOR | Golem.ExitFlags.EXIT_CLOSED;

                populateDungeon(plane.dungeon);
            };
        }
    },
};");

INSERT INTO plane_script (id, plane_id, script_id) VALUES (1, 1, 1);

CREATE INDEX index_script_name ON scripts(name);