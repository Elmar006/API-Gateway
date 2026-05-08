DROP INDEX IF EXISTS idx_logs_trace;
ALTER TABLE request_logs DROP COLUMN IF EXISTS trace_id;
