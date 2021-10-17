CREATE TABLE zones (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(255),
    `low` INT NOT NULL,
    `high` INT NOT NULL,
    `reset_message` TEXT,
    `reset_frequency` INT NOT NULL,

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
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);

CREATE TABLE exits (
    `id` BIGINT NOT NULL,

    `room_id` BIGINT NOT NULL,
    `to_room_id` BIGINT NULL,
    `direction` INT NOT NULL,
    `flags` INT NOT NULL,
    
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),
    
    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (room_id) REFERENCES rooms(id),
    FOREIGN KEY (to_room_id) REFERENCES rooms(id)
);

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
    `experience_required_modifier` FLOAT NOT NULL DEFAULT 1.0,

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
    `room_id` BIGINT NOT NULL,
    `race_id` BIGINT NOT NULL,
    `job_id` BIGINT NOT NULL,

    `level` INT NOT NULL,
    `experience` BIGINT NOT NULL,

    `practices` INT NOT NULL,

    `health` INT NOT NULL,
    `max_health` INT NOT NULL,

    `mana` INT NOT NULL,
    `max_mana` INT NOT NULL,
    
    `stamina` INT NOT NULL,
    `max_stamina` INT NOT NULL,

    `stat_str` INT NOT NULL,
    `stat_dex` INT NOT NULL,
    `stat_int` INT NOT NULL,
    `stat_wis` INT NOT NULL,
    `stat_con` INT NOT NULL,
    `stat_cha` INT NOT NULL,
    `stat_lck` INT NOT NULL,

    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (room_id) REFERENCES rooms(id),
    FOREIGN KEY (race_id) REFERENCES races(id),
    FOREIGN KEY (job_id) REFERENCES jobs(id)
);

CREATE TABLE mobiles (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(255) NOT NULL,
    `short_description` TEXT,
    `long_description` TEXT,
    `description` TEXT,

    `race_id` BIGINT NOT NULL,
    `job_id` BIGINT NOT NULL,

    `level` INT NOT NULL,
    `experience` INT NOT NULL,

    `health` INT NOT NULL,
    `max_health` INT NOT NULL,

    `mana` INT NOT NULL,
    `max_mana` INT NOT NULL,
    
    `stamina` INT NOT NULL,
    `max_stamina` INT NOT NULL,

    `stat_str` INT NOT NULL,
    `stat_dex` INT NOT NULL,
    `stat_int` INT NOT NULL,
    `stat_wis` INT NOT NULL,
    `stat_con` INT NOT NULL,
    `stat_cha` INT NOT NULL,
    `stat_lck` INT NOT NULL,
    
    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (race_id) REFERENCES races(id),
    FOREIGN KEY (job_id) REFERENCES jobs(id)
);

CREATE TABLE resets (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `zone_id` BIGINT NOT NULL,
    `room_id` BIGINT NOT NULL,

    `type` ENUM('mobile', 'room', 'object') NOT NULL,

    `value_1` INT,
    `value_2` INT,
    `value_3` INT,
    `value_4` INT,
    
    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (room_id) REFERENCES rooms(id),
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);

CREATE TABLE objects (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `zone_id` BIGINT NOT NULL,

    `name` VARCHAR(255) NOT NULL,
    `short_description` VARCHAR(255) NOT NULL,
    `long_description` VARCHAR(255) NOT NULL,
    `description` TEXT,
    `flags` INT,

    `item_type` ENUM ('protoplasm', 'light', 'potion', 'scroll', 'container', 'armor', 'weapon', 'furniture', 'sign') NOT NULL DEFAULT 'protoplasm',
    `value_1` INT,
    `value_2` INT,
    `value_3` INT,
    `value_4` INT,
    
    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (zone_id) REFERENCES zones(id)
);

CREATE TABLE object_instances (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `parent_id` BIGINT,

    `name` VARCHAR(255) NOT NULL,
    `short_description` VARCHAR(255) NOT NULL,
    `long_description` VARCHAR(255) NOT NULL,
    `description` TEXT,
    `flags` INT,

    `item_type` ENUM ('protoplasm', 'light', 'potion', 'scroll', 'container', 'armor', 'weapon', 'furniture', 'sign') NOT NULL DEFAULT 'protoplasm',

    `value_1` INT,
    `value_2` INT,
    `value_3` INT,
    `value_4` INT,
    
    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id)
);

CREATE TABLE player_character_object (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `player_character_id` BIGINT NOT NULL,
    `object_instance_id` BIGINT NOT NULL,
    
    /* Timestamps & soft deletion */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (player_character_id) REFERENCES player_characters(id),
    FOREIGN KEY (object_instance_id) REFERENCES object_instances(id)
);

/* Seed data */
INSERT INTO zones(id, name, low, high, reset_message, reset_frequency) VALUES (1, 'Limbo', 1, 8192, '{DYou hear a faint rumbling in the distance.{x', 15);

INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (1, 1, 'Limbo', 'Floating in an ethereal void, the central nexus of starlit lanes through the heavens before time.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (2, 1, 'Office of the Developer', 'An empty room with a lawnchair and a bust of Beethoven wearing sunglasses.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (3, 1, 'Featureless Corridor in Space', 'Flickering torches in the void serve as guideposts marking lanes throughout the astral void, linking discrete spaces.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (4, 1, 'Featureless Corridor in Space', 'Flickering torches in the void serve as guideposts marking lanes throughout the astral void, linking discrete spaces.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (5, 1, 'Featureless Corridor in Space', 'Flickering torches in the void serve as guideposts marking lanes throughout the astral void, linking discrete spaces.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (6, 1, 'A Cell', 'A foul stink fills the musty, stale air of a prison cell profaned by unspeakable experiments.', 0);

INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type) VALUES (1, 1, 'ball protoplasm', 'a ball of protoplasm', 'A ball of protoplasm has been left here.', 'This is some generic object entity without definition, left strewn about by an absent-minded developer!', 0, 'protoplasm');
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type) VALUES (2, 1, 'sign post signpost', 'a signpost', 'A signpost hangs in the aether beside a foreboding trapdoor.', "{YWelcome to Golem!{x\r\n\r\n{CThis pre-alpha MUD is in active development.\r\n\r\nBeneath this safe zone welcome lobby is a test dungeon with multiple floors\r\nwhose mazes are regenerated each reboot.\r\n\r\nFind updates and information on development at https://github.com/jskz/golem\r\n\r\n{WProblems? {wFeel free to {Wcontact{w the developer at {Wjames@jskarzin.org{w.{x", 0, 'sign');

/* developer office <-> limbo */
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (1, 1, 2, 0, 3);
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (2, 2, 1, 2, 3);

/* limbo <-> corridor central */
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (3, 1, 3, 2, 0);
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (4, 3, 1, 0, 0);

/* corridor central <-> corridor west */
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (5, 3, 4, 3, 0);
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (6, 4, 3, 1, 0);

/* corridor central <-> corridor east */
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (7, 3, 5, 1, 0);
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (8, 5, 3, 3, 0);

/* corridor east <-south-> monster cell A */
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (9, 5, 6, 2, 3);
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (10, 6, 5, 0, 3);

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
    jobs(id, name, display_name, experience_required_modifier, playable)
VALUES
    (1, 'warrior', 'Warrior', 1.0, 1),
    (2, 'thief', 'Thief', 1.1, 1),
    (3, 'mage', 'Mage', 1.25, 1),
    (4, 'cleric', 'Cleric', 1.5, 1);

/* Insert a testing admin character with details: Admin/password */
INSERT INTO
    player_characters(id, username, password_hash, wizard, room_id, race_id, job_id, level, experience, practices, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
VALUES
    (1, 'Admin', '$2a$10$sS5pzrKaD9qeG3ntkT7.gOohefnxSy/9OHR/p1uImyTL2edzYeJzW', 1, 1, 1, 3, 60, 0, 0, 100, 100, 100, 100, 100, 100, 18, 18, 18, 18, 18, 18, 18);

/* Test NPC in Limbo area */
INSERT INTO
    mobiles(id, name, short_description, long_description, description, race_id, job_id, level, experience, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
VALUES
    (1, 'animated animate slime', 'an animated slime', 'An animated slime languidly oozes here.', 'Despite its benign appearance, this puddle of ooze seems poised to strike.', 1, 1, 5, 1250, 15, 15, 100, 100, 100, 100, 12, 12, 12, 12, 12, 12, 10);

INSERT INTO
    resets(id, zone_id, room_id, type, value_1, value_2, value_3, value_4)
VALUES
    (1, 1, 6, 'mobile', 1, 3, 1, 1);

INSERT INTO
    resets(id, zone_id, room_id, type, value_1, value_2, value_3, value_4)
VALUES
    (2, 1, 1, 'object', 2, 1, 1, 1);

CREATE INDEX index_pc_username ON player_characters(username);
CREATE INDEX index_race_name ON races(name);
CREATE INDEX index_job_name ON jobs(name);