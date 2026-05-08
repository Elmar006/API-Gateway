-- Middlewares are reusable per-route handlers (CORS, header rewrite, custom
-- rate-limit overrides, etc.). They live in their own table so a route can
-- reference many middlewares in a defined order.
CREATE TABLE IF NOT EXISTS middlewares (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(64)  NOT NULL UNIQUE,
    kind        VARCHAR(32)  NOT NULL
        CHECK (kind IN ('cors', 'header_rewrite', 'rate_limit_override', 'request_id', 'strip_prefix')),
    config      JSONB        NOT NULL DEFAULT '{}'::jsonb,
    is_active   BOOLEAN      NOT NULL DEFAULT TRUE,
    description TEXT         NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_middlewares_kind ON middlewares(kind);
CREATE INDEX IF NOT EXISTS idx_middlewares_name ON middlewares(name);

-- Pivot: routes ↔ middlewares with explicit ordering.
CREATE TABLE IF NOT EXISTS route_middlewares (
    route_id      INT NOT NULL REFERENCES routes(id) ON DELETE CASCADE,
    middleware_id INT NOT NULL REFERENCES middlewares(id) ON DELETE CASCADE,
    sort_order    INT NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (route_id, middleware_id)
);

CREATE INDEX IF NOT EXISTS idx_route_middlewares_route ON route_middlewares(route_id, sort_order);
