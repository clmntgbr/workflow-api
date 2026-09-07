-- +goose Up
-- +goose StatementBegin
UPDATE workflows
SET
    schedule_type = 'none',
    schedule_interval_value = 0,
    schedule_interval_unit = NULL,
    schedule_at = NULL,
    next_run_at = NULL,
    status = 'active',
    updated_at = NOW()
WHERE status IN ('inactive', 'canceled');

ALTER TABLE workflows
    DROP CONSTRAINT IF EXISTS workflows_status_check;

ALTER TABLE workflows
    ALTER COLUMN status SET DEFAULT 'active';

ALTER TABLE workflows
    ADD CONSTRAINT workflows_status_check CHECK (status IN ('active', 'deleted'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE workflows
    DROP CONSTRAINT IF EXISTS workflows_status_check;

ALTER TABLE workflows
    ALTER COLUMN status SET DEFAULT 'inactive';

ALTER TABLE workflows
    ADD CONSTRAINT workflows_status_check CHECK (status IN ('active', 'inactive', 'deleted', 'canceled'));
-- +goose StatementEnd
