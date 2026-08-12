-- +goose Up
-- +goose StatementBegin
ALTER TABLE endpoints
    ADD COLUMN IF NOT EXISTS body JSONB NOT NULL DEFAULT '{}'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE endpoints
    DROP COLUMN IF EXISTS body;
-- +goose StatementEnd
