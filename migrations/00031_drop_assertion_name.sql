-- +goose Up
-- +goose StatementBegin
ALTER TABLE assertions
    DROP COLUMN IF EXISTS name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE assertions
    ADD COLUMN IF NOT EXISTS name VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd
