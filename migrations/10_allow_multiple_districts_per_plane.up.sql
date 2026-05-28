ALTER TABLE districts ADD INDEX idx_districts_plane_id (plane_id);
ALTER TABLE districts DROP INDEX plane_id;
