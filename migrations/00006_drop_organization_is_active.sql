-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_projects_is_active;
ALTER TABLE projects DROP COLUMN IF EXISTS is_active;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
CREATE INDEX IF NOT EXISTS idx_projects_is_active ON projects (is_active);
-- +goose StatementEnd
