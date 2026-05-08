-- Environments group routes by deployment target (production, staging, …).
-- Optional on routes; existing routes default to a single auto-created
-- "production" environment so the introduction is a no-op for live data.
CREATE TABLE IF NOT EXISTS environments (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(50)  NOT NULL UNIQUE,
    base_domain VARCHAR(255) NOT NULL DEFAULT '',
    color       VARCHAR(16)  NOT NULL DEFAULT '#64748b',
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_environments_name ON environments(name);

-- Bootstrap a default environment so legacy routes can be associated with it.
INSERT INTO environments (name, color, description)
VALUES ('production', '#10b981', 'Default environment for all routes')
ON CONFLICT (name) DO NOTHING;

-- Add environment_id to routes (nullable so existing rows are unaffected).
ALTER TABLE routes
    ADD COLUMN IF NOT EXISTS environment_id INT
        REFERENCES environments(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_routes_environment ON routes(environment_id);
