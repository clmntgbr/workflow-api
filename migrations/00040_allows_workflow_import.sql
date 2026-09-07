-- +goose Up
-- +goose StatementBegin
ALTER TABLE quotas
    ADD COLUMN IF NOT EXISTS allows_workflow_import BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE quotas SET allows_workflow_import = TRUE WHERE id IN (
    '01941f29-7c02-798d-b4c6-9d777505cdbc',
    '01941f29-7c03-7a99-937e-5307a3f91734'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE quotas DROP COLUMN IF EXISTS allows_workflow_import;
-- +goose StatementEnd
