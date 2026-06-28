ALTER TABLE `jobs` ADD COLUMN `health_gain_min` INT NOT NULL DEFAULT 2 CHECK (`health_gain_min` >= 0);
ALTER TABLE `jobs` ADD COLUMN `health_gain_max` INT NOT NULL DEFAULT 2 CHECK (`health_gain_max` >= 0);
ALTER TABLE `jobs` ADD COLUMN `mana_gain_divisor` INT NOT NULL DEFAULT 1 CHECK (`mana_gain_divisor` > 0);
ALTER TABLE `jobs` ADD COLUMN `stamina_gain_min` INT NOT NULL DEFAULT 1 CHECK (`stamina_gain_min` >= 0);
ALTER TABLE `jobs` ADD COLUMN `stamina_gain_max` INT NOT NULL DEFAULT 1 CHECK (`stamina_gain_max` >= 0);
ALTER TABLE `jobs` ADD COLUMN `stamina_gain_floor` INT NOT NULL DEFAULT 1 CHECK (`stamina_gain_floor` >= 0);

UPDATE
    `jobs`
SET
    `health_gain_min` = CASE `name`
        WHEN 'warrior' THEN 11
        WHEN 'thief' THEN 8
        WHEN 'mage' THEN 6
        WHEN 'cleric' THEN 7
        ELSE `health_gain_min`
    END,
    `health_gain_max` = CASE `name`
        WHEN 'warrior' THEN 15
        WHEN 'thief' THEN 13
        WHEN 'mage' THEN 8
        WHEN 'cleric' THEN 10
        ELSE `health_gain_max`
    END,
    `mana_gain_divisor` = CASE `name`
        WHEN 'warrior' THEN 2
        WHEN 'thief' THEN 2
        WHEN 'mage' THEN 1
        WHEN 'cleric' THEN 1
        ELSE `mana_gain_divisor`
    END,
    `stamina_gain_min` = CASE `name`
        WHEN 'warrior' THEN 8
        WHEN 'thief' THEN 10
        WHEN 'mage' THEN 4
        WHEN 'cleric' THEN 5
        ELSE `stamina_gain_min`
    END,
    `stamina_gain_max` = CASE `name`
        WHEN 'warrior' THEN 11
        WHEN 'thief' THEN 14
        WHEN 'mage' THEN 7
        WHEN 'cleric' THEN 8
        ELSE `stamina_gain_max`
    END,
    `stamina_gain_floor` = CASE `name`
        WHEN 'warrior' THEN 6
        WHEN 'thief' THEN 7
        WHEN 'mage' THEN 4
        WHEN 'cleric' THEN 5
        ELSE `stamina_gain_floor`
    END;
