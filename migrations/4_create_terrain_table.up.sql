CREATE TABLE terrain (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(64) NOT NULL UNIQUE,
    `glyph_colour` VARCHAR(32) NOT NULL,
    `map_glyph` VARCHAR(32) NOT NULL,
    `movement_cost` SMALLINT NOT NULL,
    `flags` INT NOT NULL DEFAULT 0,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id)
);

INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (1, 'cave-wall', '{D', '#', -1, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (2, 'cave-deep-wall-1', '', ' ', -1, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (3, 'cave-deep-wall-2', '{D', ':', -1, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (4, 'cave-deep-wall-3', '{y', '.', -1, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (5, 'cave-deep-wall-4', '{D', '.', -1, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (6, 'cave-deep-wall-5', '{y', ':', -1, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (7, 'cave-tunnel', '{c', '.', 2, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (8, 'ocean', '{B', '~', 8, 1);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (9, 'overworld-city-exterior', '{w', 'I', -1, 1);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (10, 'overworld-city-interior', '{x', ' ', -1, 1);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (11, 'overworld-city-entrance', '{W', 'O', 0, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (12, 'plains', '{G', '.', 5, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (13, 'field', '{G', ':', 5, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (14, 'shore', '{Y', ':', 10, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (15, 'shallow-water', '{C', '~', 6, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (16, 'light-forest', '{G', '*', 8, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (17, 'deep-forest', '{G', '*', 8, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (18, 'hills', '{y', '^', 16, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (19, 'mountains', '{D', '^', 20, 0);
INSERT INTO terrain(id, name, glyph_colour, map_glyph, movement_cost, flags) VALUES (20, 'mountaintop', '{W', '^', 50, 0);