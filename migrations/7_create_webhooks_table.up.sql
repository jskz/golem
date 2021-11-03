CREATE TABLE webhooks (
    `id` BIGINT NOT NULL AUTO_INCREMENT,

    `uuid` VARCHAR(255) NOT NULL UNIQUE,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id)
);

CREATE TABLE webhook_script (
    `id` BIGINT NOT NULL AUTO_INCREMENT,
    `webhook_id` BIGINT NOT NULL,
    `script_id` BIGINT NOT NULL,

    /* Timestamps */
    `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP NOT NULL DEFAULT NOW() ON UPDATE NOW(),

    PRIMARY KEY (id),
    FOREIGN KEY (webhook_id) REFERENCES webhook(id),
    FOREIGN KEY (script_id) REFERENCES scripts(id)
);