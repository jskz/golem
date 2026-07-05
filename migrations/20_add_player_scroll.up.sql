ALTER TABLE `player_characters` ADD COLUMN `scroll` INT NOT NULL DEFAULT 50 CHECK (`scroll` >= 0);
