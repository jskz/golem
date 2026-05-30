ALTER TABLE player_characters ADD COLUMN `plane_id` BIGINT NULL;
ALTER TABLE player_characters ADD COLUMN `plane_x` INT NULL;
ALTER TABLE player_characters ADD COLUMN `plane_y` INT NULL;
ALTER TABLE player_characters ADD COLUMN `plane_z` INT NULL;
CREATE INDEX `index_pc_plane_id` ON player_characters(`plane_id`);
