-- +goose Up
ALTER TABLE variables DROP COLUMN IF EXISTS last_value;
ALTER TABLE variables DROP COLUMN IF EXISTS is_secret;
ALTER TABLE variables DROP COLUMN IF EXISTS default_value;

-- +goose Down
ALTER TABLE variables ADD COLUMN IF NOT EXISTS is_secret BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE variables ADD COLUMN IF NOT EXISTS last_value JSONB;
ALTER TABLE variables ADD COLUMN IF NOT EXISTS default_value JSONB;
