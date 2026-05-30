CREATE TABLE districts (
    `id` INTEGER PRIMARY KEY,

    `plane_id` BIGINT NOT NULL,
    
    `x` INT NOT NULL,
    `y` INT NOT NULL,
    `z` INT NOT NULL,
    `width` INT NOT NULL,
    `height` INT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (plane_id) REFERENCES planes(id)
);

CREATE TABLE district_script (
    `id` INTEGER PRIMARY KEY,
    `district_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (district_id) REFERENCES districts(id) ON DELETE CASCADE,
    FOREIGN KEY (script_id) REFERENCES scripts(id) ON DELETE CASCADE
);
