-- API tokens for programmatic access alongside JWT login tokens.
-- The hash is bcrypt of the raw token string; the raw token is shown to the
-- user exactly once at creation time and cannot be recovered later.
CREATE TABLE IF NOT EXISTS api_tokens (
    id           SERIAL       PRIMARY KEY,
    user_id      INT          NOT NULL REFERENCES admin_users(id) ON DELETE CASCADE,
    name         VARCHAR(80)  NOT NULL,
    -- Short non-secret prefix shown in the UI as "token starts with…" so the
    -- user can identify which token is which without revealing the secret.
    prefix       VARCHAR(16)  NOT NULL,
    token_hash   VARCHAR(255) NOT NULL,
    -- Comma-separated scope list; validated at issue time. Wildcard "*" allows all.
    scopes       VARCHAR(512) NOT NULL DEFAULT '',
    last_used_at TIMESTAMPTZ,
    expires_at   TIMESTAMPTZ,
    revoked_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_tokens_user      ON api_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_api_tokens_prefix    ON api_tokens(prefix);
CREATE INDEX IF NOT EXISTS idx_api_tokens_active    ON api_tokens(revoked_at) WHERE revoked_at IS NULL;

-- Add updated_at to admin_users to support edit operations.
ALTER TABLE admin_users
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Tighten role check to align with the application-level role enum.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'admin_users_role_check'
    ) THEN
        ALTER TABLE admin_users
            ADD CONSTRAINT admin_users_role_check
            CHECK (role IN ('admin', 'editor', 'viewer'));
    END IF;
END$$;
