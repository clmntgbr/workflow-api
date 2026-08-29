-- +goose Up
ALTER TABLE steps
    ADD COLUMN IF NOT EXISTS expression TEXT NULL;

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_type_check CHECK (type IN ('http', 'delay', 'condition'));

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_method_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_method_check CHECK (
        (type = 'http' AND method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'))
        OR (type IN ('delay', 'condition'))
    );

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_delay_config_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_type_config_check CHECK (
        (type = 'http' AND endpoint_id IS NOT NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NULL)
        OR (type = 'delay' AND endpoint_id IS NULL AND delay_duration_seconds IS NOT NULL AND delay_duration_seconds > 0 AND expression IS NULL)
        OR (type = 'condition' AND endpoint_id IS NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NOT NULL AND btrim(expression) <> '')
    );

ALTER TABLE connections
    ADD COLUMN IF NOT EXISTS branch VARCHAR(5) NULL;

ALTER TABLE connections
    ADD CONSTRAINT connections_branch_check CHECK (branch IS NULL OR branch IN ('true', 'false'));

ALTER TABLE step_runs
    ADD COLUMN IF NOT EXISTS matched_branch BOOLEAN NULL;

ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_method_check;

ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_method_check CHECK (
        (step_type = 'http' AND method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'))
        OR (step_type IN ('delay', 'condition'))
    );

-- +goose Down
ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_method_check;

ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_method_check CHECK (
        (step_type = 'http' AND method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'))
        OR (step_type = 'delay')
    );

ALTER TABLE step_runs
    DROP COLUMN IF EXISTS matched_branch;

ALTER TABLE connections
    DROP CONSTRAINT IF EXISTS connections_branch_check;

ALTER TABLE connections
    DROP COLUMN IF EXISTS branch;

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_config_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_delay_config_check CHECK (
        (type = 'http' AND endpoint_id IS NOT NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0))
        OR (type = 'delay' AND endpoint_id IS NULL AND delay_duration_seconds IS NOT NULL AND delay_duration_seconds > 0)
    );

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_method_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_method_check CHECK (
        (type = 'http' AND method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'))
        OR (type = 'delay')
    );

ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_type_check CHECK (type IN ('http', 'delay'));

ALTER TABLE steps
    DROP COLUMN IF EXISTS expression;
