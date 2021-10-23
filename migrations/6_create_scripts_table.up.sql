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

CREATE INDEX index_script_name ON scripts(name);