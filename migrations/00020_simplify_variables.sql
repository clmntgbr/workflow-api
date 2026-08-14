-- +goose Up
ALTER TABLE variables DROP COLUMN last_value;
ALTER TABLE variables DROP COLUMN is_secret;
ALTER TABLE variables DROP COLUMN default_value;

-- +goose Down
ALTER TABLE variables ADD COLUMN is_secret BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE variables ADD COLUMN last_value JSONB;
ALTER TABLE variables ADD COLUMN default_value JSONB;
