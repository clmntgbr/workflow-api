-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS active_organization_id UUID NULL
        REFERENCES organizations (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_users_active_organization_id
    ON users (active_organization_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_active_organization_id;
ALTER TABLE users DROP COLUMN IF EXISTS active_organization_id;
-- +goose StatementEnd
