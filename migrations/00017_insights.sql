-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS insights (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    step_run_id UUID NOT NULL REFERENCES step_runs (id) ON DELETE CASCADE,
    start_time TIMESTAMPTZ NULL,
    end_time TIMESTAMPTZ NULL,
    queue_time_ns BIGINT NULL,
    dns_lookup_duration_ns BIGINT NULL,
    tcp_connection_time_ns BIGINT NULL,
    tls_handshake_time_ns BIGINT NULL,
    ttfb_ns BIGINT NULL,
    duration_ns BIGINT NULL,
    status_code INTEGER NULL,
    response_size BIGINT NULL,
    request_size BIGINT NULL,
    attempt_number INTEGER NULL,
    total_attempts INTEGER NULL,
    error_message TEXT NULL,
    error_type VARCHAR(64) NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_insights_step_run ON insights (step_run_id);
CREATE INDEX IF NOT EXISTS idx_insights_status_code ON insights (status_code);
CREATE INDEX IF NOT EXISTS idx_insights_error_type ON insights (error_type);
CREATE INDEX IF NOT EXISTS idx_insights_created_at ON insights (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS insights;
-- +goose StatementEnd
