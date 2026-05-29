ALTER TABLE player_characters
    ADD COLUMN `plane_id` BIGINT NULL AFTER `room_id`,
    ADD COLUMN `plane_x` INT NULL AFTER `plane_id`,
    ADD COLUMN `plane_y` INT NULL AFTER `plane_x`,
    ADD COLUMN `plane_z` INT NULL AFTER `plane_y`,
    ADD INDEX `index_pc_plane_id` (`plane_id`),
    ADD CONSTRAINT `fk_player_characters_plane_id` FOREIGN KEY (`plane_id`) REFERENCES planes(id);
