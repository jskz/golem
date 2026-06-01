CREATE UNIQUE INDEX IF NOT EXISTS `index_pc_username_unique` ON player_characters(username) WHERE deleted_at IS NULL;
DROP INDEX IF EXISTS `index_pc_username`;
