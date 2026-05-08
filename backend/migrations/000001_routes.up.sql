CREATE TABLE IF NOT EXISTS routes (
    id              SERIAL PRIMARY KEY,
    method          VARCHAR(10)  NOT NULL,
    path_pattern    VARCHAR(255) NOT NULL,
    target_url      VARCHAR(255) NOT NULL,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    priority        INT          NOT NULL DEFAULT 0,
    rate_limit      INT          NOT NULL DEFAULT 0,
    require_auth    BOOLEAN      NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    UNIQUE (method, path_pattern)
);

CREATE INDEX IF NOT EXISTS idx_routes_active   ON routes(is_active);
CREATE INDEX IF NOT EXISTS idx_routes_priority ON routes(priority DESC);
