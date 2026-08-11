-- +goose Up
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_organizations_is_active;
ALTER TABLE organizations DROP COLUMN IF EXISTS is_active;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
CREATE INDEX IF NOT EXISTS idx_organizations_is_active ON organizations (is_active);
-- +goose StatementEnd
