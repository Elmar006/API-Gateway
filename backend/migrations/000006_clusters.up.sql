-- Clusters group multiple upstream targets behind a single logical endpoint.
-- A route may either keep its existing single `target_url` (back-compat) or
-- reference a cluster to enable load balancing across many backends.
CREATE TABLE IF NOT EXISTS clusters (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    strategy    VARCHAR(16)  NOT NULL DEFAULT 'round_robin'
        CHECK (strategy IN ('round_robin', 'weighted', 'least_conn')),
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_clusters_name ON clusters(name);

CREATE TABLE IF NOT EXISTS cluster_targets (
    id          SERIAL PRIMARY KEY,
    cluster_id  INT          NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    url         VARCHAR(255) NOT NULL,
    weight      INT          NOT NULL DEFAULT 1 CHECK (weight >= 0),
    is_healthy  BOOLEAN      NOT NULL DEFAULT TRUE,
    last_check  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (cluster_id, url)
);

CREATE INDEX IF NOT EXISTS idx_cluster_targets_cluster ON cluster_targets(cluster_id);
CREATE INDEX IF NOT EXISTS idx_cluster_targets_health  ON cluster_targets(cluster_id, is_healthy);

-- Routes may reference a cluster instead of a single target_url.
ALTER TABLE routes
    ADD COLUMN IF NOT EXISTS cluster_id INT
        REFERENCES clusters(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_routes_cluster ON routes(cluster_id);
