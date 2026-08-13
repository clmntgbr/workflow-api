-- +goose Up
-- +goose StatementBegin
ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS schedule_type VARCHAR(20) NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS schedule_interval_value INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS schedule_interval_unit VARCHAR(20) NULL,
    ADD COLUMN IF NOT EXISTS schedule_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS schedule_timezone VARCHAR(64) NOT NULL DEFAULT 'UTC',
    ADD COLUMN IF NOT EXISTS next_run_at TIMESTAMPTZ NULL;

UPDATE workflows
SET
    schedule_type = 'recurring',
    schedule_interval_value = schedule_interval_minutes,
    schedule_interval_unit = 'minute',
    next_run_at = CASE
        WHEN status = 'active' THEN NOW() + (schedule_interval_minutes * INTERVAL '1 minute')
        ELSE NULL
    END
WHERE schedule_interval_minutes >= 10;

ALTER TABLE workflows
    DROP COLUMN IF EXISTS schedule_interval_minutes;

ALTER TABLE workflows
    ADD CONSTRAINT workflows_schedule_type_check
        CHECK (schedule_type IN ('none', 'recurring', 'once')),
    ADD CONSTRAINT workflows_schedule_interval_unit_check
        CHECK (
            schedule_interval_unit IS NULL
            OR schedule_interval_unit IN ('minute', 'hour', 'day', 'week', 'month', 'year')
        );

CREATE INDEX IF NOT EXISTS idx_workflows_next_run_at
    ON workflows (next_run_at)
    WHERE status = 'active' AND next_run_at IS NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_workflows_next_run_at;

ALTER TABLE workflows
    DROP CONSTRAINT IF EXISTS workflows_schedule_type_check,
    DROP CONSTRAINT IF EXISTS workflows_schedule_interval_unit_check;

ALTER TABLE workflows
    ADD COLUMN IF NOT EXISTS schedule_interval_minutes INTEGER NOT NULL DEFAULT 0;

UPDATE workflows
SET schedule_interval_minutes = CASE
    WHEN schedule_type = 'recurring' AND schedule_interval_unit = 'minute' THEN schedule_interval_value
    WHEN schedule_type = 'recurring' AND schedule_interval_unit = 'hour' THEN schedule_interval_value * 60
    ELSE 0
END;

ALTER TABLE workflows
    DROP COLUMN IF EXISTS next_run_at,
    DROP COLUMN IF EXISTS schedule_timezone,
    DROP COLUMN IF EXISTS schedule_at,
    DROP COLUMN IF EXISTS schedule_interval_unit,
    DROP COLUMN IF EXISTS schedule_interval_value,
    DROP COLUMN IF EXISTS schedule_type;
-- +goose StatementEnd
