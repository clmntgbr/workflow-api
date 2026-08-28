-- +goose Up
ALTER TABLE activity_logs DROP COLUMN IF EXISTS payload;

-- +goose Down
ALTER TABLE activity_logs ADD COLUMN IF NOT EXISTS payload JSONB NOT NULL DEFAULT '{}';
