-- Add OpenTelemetry trace_id to request_logs so a logged request can be
-- correlated with traces in an external tracing backend (Jaeger / Tempo /
-- any OTLP receiver). Nullable: requests are still logged when tracing is
-- disabled or the upstream did not produce a span.
ALTER TABLE request_logs
    ADD COLUMN IF NOT EXISTS trace_id VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_logs_trace ON request_logs(trace_id) WHERE trace_id IS NOT NULL;
