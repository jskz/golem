ALTER TABLE `player_characters` ADD COLUMN `trains` INT NOT NULL DEFAULT 3 CHECK (`trains` >= 0);
