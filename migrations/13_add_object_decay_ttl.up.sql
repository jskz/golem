ALTER TABLE `objects` ADD COLUMN `ttl` INT NOT NULL DEFAULT 0;
ALTER TABLE `object_instances` ADD COLUMN `ttl` INT NOT NULL DEFAULT 0;

UPDATE
    `objects`
SET
    `ttl` = 20
WHERE
    (`flags` & 8) != 0
AND
    `ttl` <= 0;

UPDATE
    `object_instances`
SET
    `ttl` = 20,
    `created_at` = CURRENT_TIMESTAMP
WHERE
    (`flags` & 8) != 0
AND
    `ttl` <= 0;
