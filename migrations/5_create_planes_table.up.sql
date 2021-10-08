CREATE TABLE planes (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL UNIQUE,
    `zone_id` INT,

    `plane_type` ENUM('void', 'maze', 'wilderness') NOT NULL DEFAULT 'void',
    `source_type` ENUM('void', 'blob', 'procedural') NOT NULL DEFAULT 'void',
    `source_value` TEXT, /* may be empty/NULL, a 2D array of terrain ids, or seed */

    `width` INT NOT NULL,
    `height` INT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (zone_id) REFERENCES zones(id),
);

INSERT INTO planes(id, name, plane_type, source_type, width, height) VALUES (1, 'limbo-maze', 'maze', 50, 50);