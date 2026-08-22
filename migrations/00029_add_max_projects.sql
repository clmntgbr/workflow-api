-- +goose Up
-- +goose StatementBegin
ALTER TABLE quotas
    ADD COLUMN IF NOT EXISTS max_projects INTEGER NOT NULL DEFAULT 1;

UPDATE quotas SET max_projects = 1, updated_at = NOW()
WHERE id = '01941f29-7c00-7558-b853-0fef65937505';

UPDATE quotas SET max_projects = 5, updated_at = NOW()
WHERE id = '01941f29-7c01-7ff3-b3eb-d5c787d32ba4';

UPDATE quotas SET max_projects = 10, updated_at = NOW()
WHERE id = '01941f29-7c02-798d-b4c6-9d777505cdbc';

UPDATE quotas SET max_projects = 50, updated_at = NOW()
WHERE id = '01941f29-7c03-7a99-937e-5307a3f91734';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE quotas DROP COLUMN IF EXISTS max_projects;
-- +goose StatementEnd
