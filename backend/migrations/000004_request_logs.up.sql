CREATE TABLE IF NOT EXISTS request_logs (
    id               BIGSERIAL PRIMARY KEY,
    request_id       UUID         NOT NULL,
    method           VARCHAR(10)  NOT NULL,
    path             VARCHAR(255) NOT NULL,
    query            TEXT,
    route_id         INT          REFERENCES routes(id) ON DELETE SET NULL,
    target_url       VARCHAR(255),
    user_id          INT,
    client_ip        INET,
    status_code      SMALLINT     NOT NULL,
    response_time_ms INT          NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_logs_created_at ON request_logs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_logs_status     ON request_logs(status_code);
CREATE INDEX IF NOT EXISTS idx_logs_route      ON request_logs(route_id);
CREATE INDEX IF NOT EXISTS idx_logs_path       ON request_logs(path);
