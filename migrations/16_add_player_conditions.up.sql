ALTER TABLE `player_characters` ADD COLUMN `condition_drunk` INT NOT NULL DEFAULT 0 CHECK (`condition_drunk` >= 0 AND `condition_drunk` <= 48);
ALTER TABLE `player_characters` ADD COLUMN `condition_full` INT NOT NULL DEFAULT 48 CHECK (`condition_full` >= 0 AND `condition_full` <= 48);
ALTER TABLE `player_characters` ADD COLUMN `condition_thirst` INT NOT NULL DEFAULT 48 CHECK (`condition_thirst` >= 0 AND `condition_thirst` <= 48);
ALTER TABLE `player_characters` ADD COLUMN `condition_hunger` INT NOT NULL DEFAULT 48 CHECK (`condition_hunger` >= 0 AND `condition_hunger` <= 48);
