CREATE TABLE terrain (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(64) NOT NULL UNIQUE,
    `map_glyph` VARCHAR(32) NOT NULL,
    `movement_cost` SMALLINT NOT NULL,
    `flags` INT NOT NULL DEFAULT 0,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id)
);

INSERT INTO terrain(id, name, map_glyph, movement_cost, flags) VALUES (1, 'cave-wall', '{D#', -1, 0);
INSERT INTO terrain(id, name, map_glyph, movement_cost, flags) VALUES (2, 'cave-deep-wall-1', ' ', -1, 0);
INSERT INTO terrain(id, name, map_glyph, movement_cost, flags) VALUES (3, 'cave-deep-wall-2', '{yx', -1, 0);
INSERT INTO terrain(id, name, map_glyph, movement_cost, flags) VALUES (4, 'cave-deep-wall-3', '{y=', -1, 0);
INSERT INTO terrain(id, name, map_glyph, movement_cost, flags) VALUES (5, 'cave-tunnel', '{c.', 2, 0);