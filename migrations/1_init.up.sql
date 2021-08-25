CREATE TABLE races (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(64) NOT NULL,
    `display_name` VARCHAR(64) NOT NULL,
    `playable` BOOLEAN NOT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE jobs (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(64) NOT NULL,
    `display_name` VARCHAR(64) NOT NULL,
    `playable` BOOLEAN NOT NULL,
    
    PRIMARY KEY (id)
);

CREATE TABLE player_characters (
    /* Identity and authentication */
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL,
    `password_hash` VARCHAR(60) NOT NULL,

    /* Gameplay fields */
    `race_id` BIGINT NOT NULL,
    `job_id` BIGINT NOT NULL,

    `health` INT NOT NULL,
    `max_health` INT NOT NULL,

    `mana` INT NOT NULL,
    `max_mana` INT NOT NULL,
    
    `stamina` INT NOT NULL,
    `max_stamina` INT NOT NULL,

    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME DEFAULT CURRENT_TIMESTAMP,

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

CREATE INDEX index_pc_username ON player_characters(username);
CREATE INDEX index_race_name ON races(name);
CREATE INDEX index_job_name ON jobs(name);