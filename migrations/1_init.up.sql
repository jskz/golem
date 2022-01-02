CREATE TABLE zones (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `name` VARCHAR(255),
    `who_description` VARCHAR(7),
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
    `id` BIGINT NOT NULL AUTO_INCREMENT,

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

    `primary_attribute` ENUM('none', 'strength', 'dexterity', 'intelligence', 'wisdom', 'constitution', 'charisma', 'luck') DEFAULT 'none',

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

    `primary_attribute` ENUM('none', 'strength', 'dexterity', 'intelligence', 'wisdom', 'constitution', 'charisma', 'luck') DEFAULT 'none',

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
    `gold` INT NOT NULL DEFAULT 0,

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

    `flags` INT NOT NULL DEFAULT 0,
    `gold` INT NOT NULL DEFAULT 0,

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

    `item_type` ENUM ('protoplasm', 'light', 'potion', 'food', 'furniture', 'drink_container', 'scroll', 'container', 'armor', 'weapon', 'sign', 'treasure', 'reagent', 'artifact', 'currency') NOT NULL DEFAULT 'protoplasm',
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
    `wear_location` INT DEFAULT -1,

    `item_type` ENUM ('protoplasm', 'light', 'potion', 'food', 'furniture', 'drink_container', 'scroll', 'container', 'armor', 'weapon', 'sign', 'treasure', 'reagent', 'artifact', 'currency') NOT NULL DEFAULT 'protoplasm',

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
INSERT INTO zones(id, name, who_description, low, high, reset_message, reset_frequency) VALUES (1, 'Limbo', 'Void', 1, 128, '{DYou hear a faint rumbling in the distance.{x', 15);

INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (1, 1, 'Limbo', 'Floating in an ethereal void, the central nexus of starlit lanes through the heavens before time.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (2, 1, 'Office of the Developer', 'An empty room with a lawnchair and a bust of Beethoven wearing sunglasses.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (3, 1, 'Featureless Corridor in Space', 'Flickering torches in the void serve as guideposts marking lanes throughout the astral void, linking discrete spaces.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (4, 1, 'Featureless Corridor in Space', 'Flickering torches in the void serve as guideposts marking lanes throughout the astral void, linking discrete spaces.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (5, 1, 'Featureless Corridor in Space', 'Flickering torches in the void serve as guideposts marking lanes throughout the astral void, linking discrete spaces.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (6, 1, 'A Cell', 'A foul stink fills the musty, stale air of a prison cell profaned by unspeakable experiments.', 0);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (7, 1, 'Training Room', 'Part library, part study, part workshop, this chamber has been set aside for all pursuits of self-mastery.', 4);
INSERT INTO rooms(id, zone_id, name, description, flags) VALUES (8, 1, 'Trading Post', "A station for commerce has somehow taken shape in the astral void.", 4);

INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (1, 1, 'ball protoplasm', 'a ball of protoplasm', 'A ball of protoplasm has been left here.', 'This is some generic object entity without definition, left strewn about by an absent-minded developer!', 0, 'protoplasm', 0, 0, 0, 0);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (2, 1, 'gold coin', 'a gold coin', 'A gold coin lies on the ground here.', 'A single gold coin.', 0, 'currency', 0, 0, 0, 0);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (3, 1, 'gold coins pile', '%d gold coins', 'There is a pile of gold coins here.', 'A pile of %d gold coins.', 0, 'currency', 0, 0, 0, 0);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (4, 1, 'sign post signpost', 'a signpost', 'A signpost hangs in the aether beside a foreboding trapdoor.', "{YWelcome to Golem!{x\r\n\r\n{CThis pre-alpha MUD is in active development.\r\n\r\nBeneath this safe zone welcome lobby is a test dungeon with multiple floors\r\nwhose mazes are regenerated each reboot.\r\n\r\nFind updates and information on development at https://github.com/jskz/golem\r\n\r\n{WProblems? {wFeel free to {Wcontact{w the developer at {Wjames@jskarzin.org{w.{x", 0, 'sign', 0, 0, 0, 0);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (5, 1, 'minor healing potion', 'a potion of minor healing', 'A hazy cyan potion lays here.', '{CTh{cis fo{Cgg{Wy c{wyan l{Wi{Cq{cu{Cid is a life-giving elixir, but the taste is not so great.{x', 1, 'potion', 0, 0, 0, 0);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (6, 1, 'swashbuckler cutlass sword', "a swashbuckler's cutlass", "A sword with a large, basket-guarded hilt was left here.", "The preferred sword of the pirate, with a pronounced basket guard at the hilt.", 7, 'weapon', 3, 6, 5, 1);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (7, 1, 'wizard wizardry pointed hat cone', 'a pointed hat', 'An unassuming cone encircled by a wide brim suggests wizardry afoot.', 'An i-conic cap perfectly fit for the magical mindspace.', 69, 'armor', 3, 3, 3, 10);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (8, 1, 'black cowl', 'a black cowl', 'A black cloth hood has been discarded here.', 'The perfect concealment for a brigand or other stalker of the night.', 69, 'armor', 1, 2, 2, 1);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (9, 1, 'mithril vest', 'a mithril vest', 'A vest of mithril links sits here.', 'A vest formed with links of mithril providing some defense against slashing weapons.', 261, 'armor', 15, 15, 15, 2);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (10, 1, 'leather boots', 'a pair of leather boots', 'Some unopinionated leather boots were left here.', 'These plain leather boots are only arguably better than nothing.', 32773, 'armor', 2, 2, 2, 1);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (11, 1, 'treasure chest secure weatherworn weather worn', 'a weatherworn treasure chest', 'A treasure chest is securely anchored here, in space.', 'A weatherworn treasure chest of unknown origin fixed by magical force within the ethereal void invites community loot-sharing.', 786432, 'container', 200, 5000, 20, 20);
INSERT INTO objects(id, zone_id, name, short_description, long_description, description, flags, item_type, value_1, value_2, value_3, value_4) VALUES (12, 1, 'belt pouch beltpouch', 'a belt pouch', 'A leather belt with a tied pouch for storage sits here.', 'A belt with a convenient pouch for holding just a couple of items.', 802817, 'container', 5, 20, 0, 0);

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

/* corridor west <-> training room */
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (11, 4, 7, 2, 3);
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (12, 7, 4, 0, 3);

/* limbo <-> trading post */
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (13, 1, 8, 1, 3);
INSERT INTO exits(id, room_id, to_room_id, direction, flags) VALUES (14, 8, 1, 3, 3);

/* Races */
INSERT INTO
    races(id, name, display_name, playable, primary_attribute)
VALUES
    (1, 'human', 'Human', 1, 'charisma'),
    (2, 'elf', 'Elf', 1, 'intelligence'),
    (3, 'dwarf', 'Dwarf', 1, 'constitution'),
    (4, 'ogre', 'Ogre', 1, 'strength');

/* Jobs */
INSERT INTO
    jobs(id, name, display_name, experience_required_modifier, playable, primary_attribute)
VALUES
    (1, 'warrior', 'Warrior', 1.0, 1, 'strength'),
    (2, 'thief', 'Thief', 1.1, 1, 'dexterity'),
    (3, 'mage', 'Mage', 1.25, 1, 'intelligence'),
    (4, 'cleric', 'Cleric', 1.5, 1, 'wisdom');

/* Insert a testing admin character with details: Admin/password */
INSERT INTO
    player_characters(id, username, password_hash, wizard, room_id, race_id, job_id, level, gold, experience, practices, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
VALUES
    (1, 'Admin', '$2a$10$MyXwV9I9wR1quCNCdX0QAuNMhFQnxlqwleyFqCI98yJs7RW/C8LDG', 1, 1, 1, 3, 60, 5000, 0, 0, 100, 100, 5000, 5000, 5000, 5000, 18, 18, 18, 18, 18, 18, 18);

/* Test NPC in Limbo area */
INSERT INTO
    mobiles(id, name, short_description, long_description, description, race_id, job_id, level, gold, experience, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
VALUES
    (1, 'animated animate slime', 'an animated slime', 'An animated slime languidly oozes here.', 'Despite its benign appearance, this puddle of ooze seems poised to strike.', 1, 1, 5, 1, 1250, 15, 15, 100, 100, 100, 100, 12, 12, 12, 12, 12, 12, 10);

INSERT INTO
    mobiles(id, name, short_description, long_description, description, flags, race_id, job_id, level, experience, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
VALUES
    (2, 'guild master guildmaster', 'the guildmaster', 'The guildmaster waits patiently to counsel and guide.', "An inviting figure waits patiently to train your fundamentals.", 32, 1, 1, 50, 1250, 15, 15, 100, 100, 100, 100, 12, 12, 12, 12, 12, 12, 10);

INSERT INTO
    mobiles(id, name, short_description, long_description, description, flags, race_id, job_id, level, experience, health, max_health, mana, max_mana, stamina, max_stamina, stat_str, stat_dex, stat_int, stat_wis, stat_con, stat_cha, stat_lck)
VALUES
    (3, 'astral shop keeper being translucent', 'the astral shop keeper', 'A translucent being tends to its interstellar store.', 'A lucent being of star-stuff and singular intent who buys and sells items.', 128, 1, 1, 50, 1250, 15, 15, 100, 100, 100, 100, 12, 12, 12, 12, 12, 12, 10);

INSERT INTO
    resets(id, zone_id, room_id, type, value_1, value_2, value_3, value_4)
VALUES
    (1, 1, 6, 'mobile', 1, 3, 1, 1);
INSERT INTO
    resets(id, zone_id, room_id, type, value_1, value_2, value_3, value_4)
VALUES
    (2, 1, 7, 'mobile', 2, 3, 1, 1);
INSERT INTO
    resets(id, zone_id, room_id, type, value_1, value_2, value_3, value_4)
VALUES
    (3, 1, 8, 'mobile', 3, 3, 1, 1);

INSERT INTO
    resets(id, zone_id, room_id, type, value_1, value_2, value_3, value_4)
VALUES
    (4, 1, 1, 'object', 4, 1, 1, 1);
INSERT INTO
    resets(id, zone_id, room_id, type, value_1, value_2, value_3, value_4)
VALUES
    (5, 1, 1, 'object', 11, 1, 1, 1);

CREATE INDEX index_pc_username ON player_characters(username);
CREATE INDEX index_race_name ON races(name);
CREATE INDEX index_job_name ON jobs(name);