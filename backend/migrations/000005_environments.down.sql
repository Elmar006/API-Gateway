DROP INDEX IF EXISTS idx_routes_environment;
ALTER TABLE routes DROP COLUMN IF EXISTS environment_id;
DROP TABLE IF EXISTS environments;
