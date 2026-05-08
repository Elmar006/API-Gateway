ALTER TABLE admin_users DROP CONSTRAINT IF EXISTS admin_users_role_check;
ALTER TABLE admin_users DROP COLUMN IF EXISTS updated_at;

DROP INDEX IF EXISTS idx_api_tokens_active;
DROP INDEX IF EXISTS idx_api_tokens_prefix;
DROP INDEX IF EXISTS idx_api_tokens_user;
DROP TABLE IF EXISTS api_tokens;
