ALTER TABLE `objects` ADD COLUMN `weight` REAL NOT NULL DEFAULT 0 CHECK (`weight` >= 0);
ALTER TABLE `object_instances` ADD COLUMN `weight` REAL NOT NULL DEFAULT 0 CHECK (`weight` >= 0);

UPDATE
    `objects`
SET
    `weight` = CASE `id`
        WHEN 5 THEN 0.5
        WHEN 6 THEN 3.0
        WHEN 7 THEN 1.0
        WHEN 8 THEN 0.5
        WHEN 9 THEN 12.0
        WHEN 10 THEN 2.0
        WHEN 11 THEN 50.0
        WHEN 12 THEN 1.0
        ELSE `weight`
    END;

UPDATE
    `object_instances`
SET
    `weight` = COALESCE(
        (
            SELECT
                `objects`.`weight`
            FROM
                `objects`
            WHERE
                `objects`.`id` = `object_instances`.`parent_id`
        ),
        0
    );
