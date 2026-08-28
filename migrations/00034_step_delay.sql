-- +goose Up
ALTER TABLE steps
    ADD COLUMN IF NOT EXISTS type VARCHAR(20) NOT NULL DEFAULT 'http',
    ADD COLUMN IF NOT EXISTS delay_duration_seconds INTEGER NULL;

ALTER TABLE steps
    ALTER COLUMN endpoint_id DROP NOT NULL;

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_method_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_type_check CHECK (type IN ('http', 'delay'));

ALTER TABLE steps
    ADD CONSTRAINT steps_method_check CHECK (
        (type = 'http' AND method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'))
        OR (type = 'delay')
    );

ALTER TABLE steps
    ADD CONSTRAINT steps_delay_config_check CHECK (
        (type = 'http' AND endpoint_id IS NOT NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0))
        OR (type = 'delay' AND endpoint_id IS NULL AND delay_duration_seconds IS NOT NULL AND delay_duration_seconds > 0)
    );

ALTER TABLE step_runs
    ADD COLUMN IF NOT EXISTS step_type VARCHAR(20) NOT NULL DEFAULT 'http',
    ADD COLUMN IF NOT EXISTS delay_duration_seconds INTEGER NULL,
    ADD COLUMN IF NOT EXISTS resume_at TIMESTAMPTZ NULL;

ALTER TABLE step_runs
    ALTER COLUMN endpoint_id DROP NOT NULL;

ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_method_check;

ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_status_check;

ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_method_check CHECK (
        (step_type = 'http' AND method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'))
        OR (step_type = 'delay')
    );

ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_status_check CHECK (
        status IN ('pending', 'running', 'waiting', 'success', 'failed', 'cancelled', 'skipped')
    );

CREATE INDEX IF NOT EXISTS idx_step_runs_waiting_resume_at
    ON step_runs (resume_at)
    WHERE status = 'waiting';

-- +goose Down
DROP INDEX IF EXISTS idx_step_runs_waiting_resume_at;

ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_status_check;

ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_status_check
        CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled', 'skipped'));

ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_method_check;

ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_method_check CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'));

ALTER TABLE step_runs
    DROP COLUMN IF EXISTS resume_at,
    DROP COLUMN IF EXISTS delay_duration_seconds,
    DROP COLUMN IF EXISTS step_type;

ALTER TABLE step_runs
    ALTER COLUMN endpoint_id SET NOT NULL;

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_delay_config_check;

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_method_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_method_check CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'));

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_check;

ALTER TABLE steps
    DROP COLUMN IF EXISTS delay_duration_seconds,
    DROP COLUMN IF EXISTS type;

ALTER TABLE steps
    ALTER COLUMN endpoint_id SET NOT NULL;
