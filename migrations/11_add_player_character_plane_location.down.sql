ALTER TABLE player_characters DROP FOREIGN KEY `fk_player_characters_plane_id`;
DROP INDEX `index_pc_plane_id` ON player_characters;
ALTER TABLE player_characters
    DROP COLUMN `plane_z`,
    DROP COLUMN `plane_y`,
    DROP COLUMN `plane_x`,
    DROP COLUMN `plane_id`;
