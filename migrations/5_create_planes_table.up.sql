CREATE TABLE planes (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `zone_id` BIGINT NOT NULL,
    `name` VARCHAR(255) NOT NULL UNIQUE,

    `plane_type` ENUM('void', 'maze', 'wilderness') NOT NULL DEFAULT 'void',
    `source_type` ENUM('void', 'blob', 'procedural') NOT NULL DEFAULT 'void',
    `source_value` TEXT, /* may be empty/NULL, a 2D array of terrain ids, or seed */

    `width` INT NOT NULL,
    `height` INT NOT NULL,
    `depth` INT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);

INSERT INTO planes(id, zone_id, name, plane_type, source_type, width, height, depth) VALUES (1, 1, 'limbo-maze', 'maze', 'procedural', 32, 32, 4);
INSERT INTO planes(id, zone_id, name, plane_type, source_type, width, height, depth) VALUES (2, 0, 'overworld', 'wilderness', 'blob', 512, 512, 1);