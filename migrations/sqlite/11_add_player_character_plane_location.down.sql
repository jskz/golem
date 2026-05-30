DROP INDEX IF EXISTS `index_pc_plane_id`;
ALTER TABLE player_characters DROP COLUMN `plane_z`;
ALTER TABLE player_characters DROP COLUMN `plane_y`;
ALTER TABLE player_characters DROP COLUMN `plane_x`;
ALTER TABLE player_characters DROP COLUMN `plane_id`;
