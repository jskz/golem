CREATE TABLE socials (
    `id` INTEGER PRIMARY KEY,

    `name` VARCHAR(255) NOT NULL UNIQUE,
    `char_no_arg` TEXT NOT NULL,
    `others_no_arg` TEXT NOT NULL,
    `char_found` TEXT NOT NULL,
    `others_found` TEXT NOT NULL,
    `vict_found` TEXT NOT NULL,
    `char_not_found` TEXT NOT NULL,
    `char_auto` TEXT NOT NULL,
    `others_auto` TEXT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    `deleted_at` TIMESTAMP NULL DEFAULT NULL,
    `deleted_by` BIGINT DEFAULT NULL
);

INSERT INTO socials (
    id,
    name,
    char_no_arg,
    others_no_arg,
    char_found,
    others_found,
    vict_found,
    char_not_found,
    char_auto,
    others_auto
) VALUES (
    1,
    'grin',
    'You grin evilly.',
    '$n grins evilly.',
    'You grin evilly at $N.',
    '$n grins evilly at $N.',
    '$n grins evilly at you.  Hmmm.  Better keep your distance.',
    'You must be delirious.',
    'You grin at yourself.  You must be getting very bad thoughts.',
    '$n grins at themself.  You must wonder what''s on their mind.'
);

INSERT INTO socials (
    id,
    name,
    char_no_arg,
    others_no_arg,
    char_found,
    others_found,
    vict_found,
    char_not_found,
    char_auto,
    others_auto
) VALUES (
    2,
    'laugh',
    'You fall down laughing.',
    '$n falls down laughing.',
    'You laugh at $N mercilessly.',
    '$n laughs at $N mercilessly.',
    '$n laughs at you mercilessly.  Hmmmmph.',
    'You can''t find the butt of your joke.',
    'You laugh at yourself.  I would, too.',
    '$n laughs at themself.  Let''s all join in!!!'
);

INSERT INTO socials (
    id,
    name,
    char_no_arg,
    others_no_arg,
    char_found,
    others_found,
    vict_found,
    char_not_found,
    char_auto,
    others_auto
) VALUES (
    3,
    'nod',
    'You nod.',
    '$n nods.',
    'You nod at $N.',
    '$n nods at $N.',
    '$n nods at you in agreement.',
    'Nod your head off -- they aren''t here.',
    'You attempt to nod at yourself and get dizzy instead.',
    '$n nods quietly to themself.  What a wacko.'
);

CREATE INDEX index_social_name ON socials(name);
