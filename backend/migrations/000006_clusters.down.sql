DROP INDEX IF EXISTS idx_routes_cluster;
ALTER TABLE routes DROP COLUMN IF EXISTS cluster_id;

DROP INDEX IF EXISTS idx_cluster_targets_health;
DROP INDEX IF EXISTS idx_cluster_targets_cluster;
DROP TABLE IF EXISTS cluster_targets;

DROP INDEX IF EXISTS idx_clusters_name;
DROP TABLE IF EXISTS clusters;
