CREATE INDEX IF NOT EXISTS `index_pc_username` ON player_characters(username);
DROP INDEX IF EXISTS `index_pc_username_unique`;
