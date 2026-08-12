-- +goose Up
-- +goose StatementBegin
ALTER TABLE endpoints
    ADD COLUMN IF NOT EXISTS query_params JSONB NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS timeout_ms INTEGER NOT NULL DEFAULT 30000,
    ADD COLUMN IF NOT EXISTS retry_on_failure BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS retry_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS retry_delay_ms INTEGER NOT NULL DEFAULT 1000;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE endpoints
    DROP COLUMN IF EXISTS retry_delay_ms,
    DROP COLUMN IF EXISTS retry_count,
    DROP COLUMN IF EXISTS retry_on_failure,
    DROP COLUMN IF EXISTS timeout_ms,
    DROP COLUMN IF EXISTS query_params;
-- +goose StatementEnd
