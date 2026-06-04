UPDATE
    `player_characters`
SET
    `max_health` = 1
WHERE
    `max_health` <= 0;

UPDATE
    `player_characters`
SET
    `max_mana` = 1
WHERE
    `max_mana` <= 0;

UPDATE
    `player_characters`
SET
    `max_stamina` = 1
WHERE
    `max_stamina` <= 0;

UPDATE
    `mobiles`
SET
    `max_health` = 1
WHERE
    `max_health` <= 0;

UPDATE
    `mobiles`
SET
    `max_mana` = 1
WHERE
    `max_mana` <= 0;

UPDATE
    `mobiles`
SET
    `max_stamina` = 1
WHERE
    `max_stamina` <= 0;

CREATE TRIGGER `player_characters_positive_resource_max_insert`
BEFORE INSERT ON `player_characters`
FOR EACH ROW
WHEN
    NEW.`max_health` <= 0
OR
    NEW.`max_mana` <= 0
OR
    NEW.`max_stamina` <= 0
BEGIN
    SELECT RAISE(ABORT, 'player character max resources must be positive');
END;

CREATE TRIGGER `player_characters_positive_resource_max_update`
BEFORE UPDATE OF `max_health`, `max_mana`, `max_stamina` ON `player_characters`
FOR EACH ROW
WHEN
    NEW.`max_health` <= 0
OR
    NEW.`max_mana` <= 0
OR
    NEW.`max_stamina` <= 0
BEGIN
    SELECT RAISE(ABORT, 'player character max resources must be positive');
END;

CREATE TRIGGER `mobiles_positive_resource_max_insert`
BEFORE INSERT ON `mobiles`
FOR EACH ROW
WHEN
    NEW.`max_health` <= 0
OR
    NEW.`max_mana` <= 0
OR
    NEW.`max_stamina` <= 0
BEGIN
    SELECT RAISE(ABORT, 'mobile max resources must be positive');
END;

CREATE TRIGGER `mobiles_positive_resource_max_update`
BEFORE UPDATE OF `max_health`, `max_mana`, `max_stamina` ON `mobiles`
FOR EACH ROW
WHEN
    NEW.`max_health` <= 0
OR
    NEW.`max_mana` <= 0
OR
    NEW.`max_stamina` <= 0
BEGIN
    SELECT RAISE(ABORT, 'mobile max resources must be positive');
END;
