CREATE TABLE player_characters (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `username` VARCHAR(64) NOT NULL,
    `password_hash` VARCHAR(60) NOT NULL,

    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` DATETIME NULL ON UPDATE CURRENT_TIMESTAMP,
    
    `deleted_at` DATETIME NULL,
    `deleted_by` BIGINT NULL,

    PRIMARY KEY (id)
);