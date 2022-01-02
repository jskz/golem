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
    FOREIGN KEY (mobile_id) REFERENCES mobiles(id) ON DELETE CASCADE,
    FOREIGN KEY (script_id) REFERENCES scripts(id) ON DELETE CASCADE
);

CREATE TABLE object_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `object_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (`object_id`) REFERENCES objects(id) ON DELETE CASCADE,
    FOREIGN KEY (script_id) REFERENCES scripts(id) ON DELETE CASCADE
);

CREATE TABLE room_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `room_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (script_id) REFERENCES scripts(id) ON DELETE CASCADE
);

CREATE TABLE plane_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `plane_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (plane_id) REFERENCES planes(id) ON DELETE CASCADE,
    FOREIGN KEY (script_id) REFERENCES scripts(id) ON DELETE CASCADE
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

                        /* Do not spawn anything at the entrance of the dungeon */
                        if(z === 0 && x == dungeon.floors[z].entryX && y == dungeon.floors[z].entryY) {
                            continue;
                        }

                        if (!cell.wall && cell.room) {
                            cell.room.flags = Golem.RoomFlags.ROOM_VIRTUAL | Golem.RoomFlags.ROOM_DUNGEON;

                            if(z > 1) {
                                cell.room.flags |= Golem.RoomFlags.ROOM_EVIL_AURA;
                            }

                            const chanceToSpawnCreature = ~~(
                                Math.random() * 100
                            );

                            if (chanceToSpawnCreature > 90) {
                                const baseMobile =
                                    Golem.game.loadMobileIndex(1);

                                try {
                                    if (baseMobile) {
                                        baseMobile.name =
                                            'aggressive slime angry agitate agitated';
                                        baseMobile.shortDescription =
                                            'an agitated slime';
                                        baseMobile.longDescription =
                                            'An angry slime has festered to the point of open hostility in the dungeon.';
                                        baseMobile.description =
                                            'This angry slime has one big chip on its gelatinous shoulder.';

                                        if(z >= 4) {
                                            baseMobile.name = 'ancient evil slime';
                                            baseMobile.shortDescription = 'an ancient, evil slime';
                                            baseMobile.longDescription = 'An ancient, evil slime guards the inner sanctum of the dungeon.';
                                            baseMobile.description = 'This putrid puddle is not going to take it anymore.';
                                        }

                                        baseMobile.level = 10 * (z + 1);
                                        baseMobile.stats[Golem.StatTypes.STAT_DEXTERITY] = 15 + (2 * (z + 1));
                                        baseMobile.health = 100 + (100 * (z * 10));
                                        baseMobile.maxHealth = 100 + (100 * (z * 10));
                                        baseMobile.stats[Golem.StatTypes.STAT_STRENGTH] = 20 + (2 * (z + 1));
                                        baseMobile.experience = 4000 + (2000 * (z + 1));

                                        baseMobile.flags =
                                            Golem.CharacterFlags.CHAR_AGGRESSIVE;

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
                limbo,
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

                // Tie the dungeon's entrance to limbo and unlock the trapdoor
                limbo.exit[Golem.Directions.DirectionDown].to = dungeonEntrance;
                dungeonEntrance.exit[Golem.Directions.DirectionUp] =
                    Golem.NewExit(
                        dungeonEntrance,
                        Golem.Directions.DirectionUp,
                        limbo,
                        Golem.ExitFlags.EXIT_IS_DOOR |
                            Golem.ExitFlags.EXIT_CLOSED
                    );
                limbo.exit[Golem.Directions.DirectionDown].flags =
                    Golem.ExitFlags.EXIT_IS_DOOR | Golem.ExitFlags.EXIT_CLOSED;

                populateDungeon(plane.dungeon);
                dungeonEntrance.flags =
                    Golem.RoomFlags.ROOM_VIRTUAL | Golem.RoomFlags.ROOM_DUNGEON | Golem.RoomFlags.ROOM_SAFE;
            };
        }
    },
};");

INSERT INTO plane_script (id, plane_id, script_id) VALUES (1, 1, 1);

INSERT INTO scripts (id, name, script) VALUES (2, 'overworld', 'function drawFilledRect(d, x, y, w, h, t, f) {
  let j, s;

  for (j = y; j < y + h; j++) {
    for (s = x; s < x + w; s++) {
      d[j][s] = f;
    }
  }

  for (s = x; s <= x + w; s++) {
    d[y][s] = t;
    d[y + h][s] = t;
  }

  for (j = y; j < y + h; j++) {
    d[j][x] = t;
    d[j][x + w] = t;
  }
}

module.exports = {
  onGenerationComplete: function (p) {
    try {
        const BUILDING_POSITION = [49, 418],
        BUILDING_WIDTH = 12,
        BUILDING_HEIGHT = 6;
        const w = p.width;
        const h = p.height;
        const terrain = p.map.layers[0].terrain;

        for (let y = 0; y < h; y++) {
            for (let x = 0; x < w; x++) {
                let nx = x/w - 0.5, 
                    ny = y/h - 0.5;
                let n = Golem.util.perlin2D(nx * 1.25, ny * 1.25);

                n += 1.0;
                n /= 2.0;

                let t = Golem.TerrainTypes.TerrainTypeOcean;

                if(n > 0.47) {
                    t = Golem.TerrainTypes.TerrainTypeShallowWater;
                }
                if (n > 0.485) {
                    t = Golem.TerrainTypes.TerrainTypeShore;
                }
                if (n > 0.50) {
                    t = Golem.TerrainTypes.TerrainTypePlains;
                }
                if (n > 0.53) {
                    t = Golem.TerrainTypes.TerrainTypeField;
                }
                if (n > 0.60) {
                    t = Golem.TerrainTypes.TerrainTypeLightForest;
                }
                if (n > 0.67) {
                    t = Golem.TerrainTypes.TerrainTypeDenseForest;
                }
                if (n > 0.73) {
                    t = Golem.TerrainTypes.TerrainTypeHills;
                }
                if (n > 0.85) {
                    t = Golem.TerrainTypes.TerrainTypeMountains;
                }
                if (n > 1.1) {
                    t = Golem.TerrainTypes.TerrainTypeSnowcappedMountains;
                }

                terrain[y][x] = t;
            }
        }

        // Create a structure representing the developer area
        drawFilledRect(
        terrain,
        BUILDING_POSITION[0],
        BUILDING_POSITION[1],
        BUILDING_WIDTH,
        BUILDING_HEIGHT,
        Golem.TerrainTypes.OverworldCityExterior,
        Golem.TerrainTypes.OverworldCityInterior
        );

        // Create the entrance
        const BOTTOM_WALL_Y = BUILDING_POSITION[1] + BUILDING_HEIGHT;
        const BOTTOM_ENTRANCE_FRONT_X = BUILDING_POSITION[0] + 3;
        const BOTTOM_ENTRANCE_FRONT_Y = BOTTOM_WALL_Y + 1;
        
        terrain[BOTTOM_WALL_Y][BOTTOM_ENTRANCE_FRONT_X] = Golem.TerrainTypes.OverworldCityEntrance;
        const templeFront = p.materializeRoom(BOTTOM_ENTRANCE_FRONT_X, BOTTOM_ENTRANCE_FRONT_Y, 0, true);     

        const foyer = Golem.game.loadRoomIndex(3);

        foyer.exit[Golem.Directions.DirectionSouth] =
            Golem.NewExit(
                foyer,
                Golem.Directions.DirectionSouth,
                templeFront,
                Golem.ExitFlags.EXIT_IS_DOOR |
                    Golem.ExitFlags.EXIT_CLOSED
            );
        templeFront.exit[Golem.Directions.DirectionNorth].to = foyer;
        templeFront.exit[Golem.Directions.DirectionNorth].flags = Golem.ExitFlags.EXIT_IS_DOOR | Golem.ExitFlags.EXIT_CLOSED;
    } catch(err) {
      println(err.toString());
    }
  },
};');

INSERT INTO plane_script (id, plane_id, script_id) VALUES (2, 2, 2);

INSERT INTO
    scripts (id, name, script) VALUES (3, 'minor-healing-potion', 'module.exports = {
    onUse: function(ch) {
        if(!ch.isEqual(this.carriedBy)) {
            ch.send("You aren\'t carrying that.\\r\\n");
            return;
        }

        ch.detachObject(this);
        ch.send("{WYou quaff " + this.getShortDescription(ch) + "{x.\\r\\n");
        ch.removeObject(this);

        const amount = ~~(Math.random() * 5) + 5;
        Golem.game.damage(null, ch, false, -amount, Golem.Combat.DamageTypeExotic);
        ch.send("{WYou feel a little bit better.{x\\r\\n");
    }
};');

INSERT INTO object_script (id, `object_id`, script_id) VALUES (1, 5, 3);

CREATE INDEX index_script_name ON scripts(name);