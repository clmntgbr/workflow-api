-- +goose Up
ALTER TABLE connections
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS created_at;

-- +goose Down
ALTER TABLE connections
    ADD COLUMN status VARCHAR(50) NOT NULL DEFAULT 'active',
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
