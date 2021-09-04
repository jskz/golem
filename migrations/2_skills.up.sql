CREATE TABLE skills (
    `id` BIGINT NOT NULL,

    `name` VARCHAR(255) NOT NULL UNIQUE,
    `type` ENUM('skill', 'spell', 'passive'),

    PRIMARY KEY (id)
);

CREATE TABLE job_skill (
    `id` BIGINT NOT NULL,

    `job_id` BIGINT NOT NULL,
    `skill_id` BIGINT NOT NULL,
    `level` INT NOT NULL,

    `complexity` BIGINT NOT NULL,
    `cost` BIGINT NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (job_id) REFERENCES jobs(id),
    FOREIGN KEY (skill_id) REFERENCES skills(id)
);

CREATE TABLE pc_skill_proficiency (
    `id` BIGINT NOT NULL,

    `player_character_id` BIGINT NOT NULL,
    `skill_id` BIGINT NOT NULL,

    `proficiency` INT NOT NULL,

    PRIMARY KEY (id),
    FOREIGN KEY (player_character_id) REFERENCES player_characters(id),
    FOREIGN KEY (skill_id) REFERENCES skills(id)
);

INSERT INTO skills(id, name, type) VALUES (1, 'dodge', 'passive');
INSERT INTO skills(id, name, type) VALUES (2, 'unarmed combat', 'passive');
INSERT INTO skills(id, name, type) VALUES (3, 'peek', 'passive');

/* Grant unarmed combat as a seed skill for all four base jobs with varying complexity and cost */
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (1, 1, 2, 1, 1, 1);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (2, 1, 2, 2, 2, 2);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (3, 1, 2, 3, 5, 5);
INSERT INTO job_skill(id, job_id, skill_id, level, complexity, cost) VALUES (4, 1, 2, 4, 5, 5);

/* Grant it mastered to the seed admin user as well */
INSERT INTO pc_skill_proficiency(id, player_character_id, skill_id, proficiency) VALUES (1, 1, 2, 100);

CREATE INDEX index_skill_name ON skills(name);