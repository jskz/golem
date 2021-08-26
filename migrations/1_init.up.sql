CREATE TABLE zones (
    `id` BIGINT NOT NULL,

    `name` VARCHAR(255),
    `low` INT NOT NULL,
    `high` INT NOT NULL,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE rooms (
    `id` BIGINT NOT NULL,
    `zone_id` BIGINT NOT NULL,

    `name` VARCHAR(255),
    `description` TEXT,
    `flags` INT,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (zone_id) REFERENCES zones(id);
);

CREATE TABLE exits (
    `id` BIGINT NOT NULL,

    `room_id` BIGINT NOT NULL,
    `to_room_id` BIGINT NULL,
    `direction` INT NOT NULL,
    `flags` INT NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (room_id) REFERENCES rooms(id),
    FOREIGN KEY (to_room_id) REFERENCES rooms(id)
);

INSERT INTO zones(id, name, low, high) VALUES (1, 'Limbo', 1, 99);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (1, 1, 'Limbo', 'Floating in an ethereal void.', 0);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (2, 1, 'Developer Room', 'Another testing room.', 0);

INSERT INTO exits(id, room_id, to_room_id, direction, flags,) VALUES (1, 1, 2, 0, 0);
INSERT INTO exits(id, room_id, to_room_id, direction, flags,) VALUES (1, 2, 1, 2, 0);

CREATE TABLE races (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(64) NOT NULL,
    `display_name` VARCHAR(64) NOT NULL,
    `playable` BOOLEAN NOT NULL,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE jobs (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(64) NOT NULL,
    `display_name` VARCHAR(64) NOT NULL,
    `playable` BOOLEAN NOT NULL,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,
    
    PRIMARY KEY (id)
);

CREATE TABLE player_characters (
    /* Identity and authentication */
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL,
    `password_hash` VARCHAR(60) NOT NULL,

    /* Admin status */
    `wizard` BOOLEAN NOT NULL,

    /* Gameplay fields */
    `race_id` BIGINT NOT NULL,
    `job_id` BIGINT NOT NULL,

    `level` INT NOT NULL,

    `health` INT NOT NULL,
    `max_health` INT NOT NULL,

    `mana` INT NOT NULL,
    `max_mana` INT NOT NULL,
    
    `stamina` INT NOT NULL,
    `max_stamina` INT NOT NULL,

    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (race_id) REFERENCES races(id),
    FOREIGN KEY (job_id) REFERENCES jobs(id)
);

/* Races */
INSERT INTO
    races(id, name, display_name, playable)
VALUES
    (1, 'human', 'Human', 1),
    (2, 'elf', 'Elf', 1),
    (3, 'dwarf', 'Dwarf', 1),
    (4, 'ogre', 'Ogre', 1);

/* Jobs */
INSERT INTO
    jobs(id, name, display_name, playable)
VALUES
    (1, 'warrior', 'Warrior', 1),
    (2, 'thief', 'Thief', 1),
    (3, 'mage', 'Mage', 1),
    (4, 'cleric', 'Cleric', 1);

/* Insert an admin character with name password */
INSERT INTO
    player_characters(id, username, password_hash, wizard, race_id, job_id, level, health, max_health, mana, max_mana, stamina, max_stamina)
VALUES
    (1, 'Admin', '$2a$10$sS5pzrKaD9qeG3ntkT7.gOohefnxSy/9OHR/p1uImyTL2edzYeJzW', 1, 1, 1, 60, 100, 100, 100, 100, 100, 100);

CREATE INDEX index_pc_username ON player_characters(username);
CREATE INDEX index_race_name ON races(name);
CREATE INDEX index_job_name ON jobs(name);