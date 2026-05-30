CREATE TABLE planes (
    `id` INTEGER PRIMARY KEY,
    `zone_id` BIGINT NOT NULL,
    `name` VARCHAR(255) NOT NULL UNIQUE,

    `plane_type` TEXT NOT NULL DEFAULT 'void',
    `source_type` TEXT NOT NULL DEFAULT 'void',
    `source_value` BLOB, /* may be empty/NULL, a 2D array of terrain ids, or seed */

    `width` INT NOT NULL,
    `height` INT NOT NULL,
    `depth` INT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO planes(id, zone_id, name, plane_type, source_type, width, height, depth) VALUES (1, 1, 'limbo-maze', 'maze', 'procedural', 32, 32, 4);
INSERT INTO planes(id, zone_id, name, plane_type, source_type, width, height, depth) VALUES (2, 0, 'overworld', 'wilderness', 'blob', 512, 512, 1);