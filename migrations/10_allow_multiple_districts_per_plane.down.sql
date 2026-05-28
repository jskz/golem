ALTER TABLE districts ADD UNIQUE INDEX plane_id (plane_id);
ALTER TABLE districts DROP INDEX idx_districts_plane_id;
