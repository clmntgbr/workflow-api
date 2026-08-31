-- +goose Up
-- +goose StatementBegin
ALTER TABLE quotas
    ADD COLUMN IF NOT EXISTS max_assertions_per_workflow INTEGER NOT NULL DEFAULT 0;

UPDATE quotas SET max_assertions_per_workflow = 5 WHERE id = '01941f29-7c00-7558-b853-0fef65937505';
UPDATE quotas SET max_assertions_per_workflow = 50 WHERE id = '01941f29-7c01-7ff3-b3eb-d5c787d32ba4';
UPDATE quotas SET max_assertions_per_workflow = 100 WHERE id = '01941f29-7c02-798d-b4c6-9d777505cdbc';
UPDATE quotas SET max_assertions_per_workflow = 200 WHERE id = '01941f29-7c03-7a99-937e-5307a3f91734';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE quotas DROP COLUMN IF EXISTS max_assertions_per_workflow;
-- +goose StatementEnd
