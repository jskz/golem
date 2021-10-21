CREATE TABLE skills (
    `id` BIGINT NOT NULL,

    `name` VARCHAR(255) NOT NULL UNIQUE,
    `type` ENUM('skill', 'spell', 'passive'),
    
    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id)
);

CREATE TABLE job_skill (
    `id` BIGINT NOT NULL,

    `job_id` BIGINT NOT NULL,
    `skill_id` BIGINT NOT NULL,
    `level` INT NOT NULL,

    `complexity` BIGINT NOT NULL,
    `cost` BIGINT NOT NULL,
    
    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (job_id) REFERENCES jobs(id),
    FOREIGN KEY (skill_id) REFERENCES skills(id)
);

CREATE TABLE pc_skill_proficiency (
    `id` BIGINT NOT NULL,

    `player_character_id` BIGINT NOT NULL,
    `skill_id` BIGINT NOT NULL,
    `job_id` BIGINT NOT NULL,

    `proficiency` INT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (player_character_id) REFERENCES player_characters(id),
    FOREIGN KEY (job_id) REFERENCES jobs(id),
    FOREIGN KEY (skill_id) REFERENCES skills(id)
);

INSERT INTO skills(id, name, type) VALUES (1, 'dodge', 'passive');
INSERT INTO skills(id, name, type) VALUES (2, 'unarmed combat', 'passive');
INSERT INTO skills(id, name, type) VALUES (3, 'peek', 'passive');
INSERT INTO skills(id, name, type) VALUES (4, 'armor', 'spell');
INSERT INTO skills(id, name, type) VALUES (5, 'fireball', 'spell');
INSERT INTO skills(id, name, type) VALUES (6, 'bash', 'skill');
INSERT INTO skills(id, name, type) VALUES (7, 'cure light', 'spell');
INSERT INTO skills(id, name, type) VALUES (8, 'magic map', 'spell');

/* Grant unarmed combat as a seed skill for all four base jobs with varying complexity and cost */
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (1, 1, 2, 1, 1, 1);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (2, 2, 2, 2, 2, 2);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (3, 3, 2, 3, 5, 5);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (4, 4, 2, 4, 5, 5);

/* Warrior defaults: bash, dodge */
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (5, 1, 6, 1, 5, 20);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (11, 1, 1, 20, 10, 10);

/* Thief defaults: peek, dodge */
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (6, 2, 3, 1, 5, 5);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (12, 2, 1, 5, 5, 5);

/* Cleric defaults: armor, cure light, magic map */
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (7, 4, 4, 1, 1, 20);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (9, 4, 7, 1, 1, 20);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (10, 4, 8, 5, 1, 20);

/* Mage defaults: fireball */
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (8, 3, 5, 1, 1, 20);

/* Grant some skills mastered to the seed admin user as well */
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (1, 1, 1, 2, 100);
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (2, 1, 2, 1, 100);
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (3, 1, 3, 2, 100);
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (4, 1, 4, 4, 100);
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (5, 1, 5, 3, 100);
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (6, 1, 6, 1, 100);
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (7, 1, 7, 4, 100);
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, job_id, proficiency) VALUES (8, 1, 8, 4, 100);

CREATE INDEX index_skill_name ON skills(name);