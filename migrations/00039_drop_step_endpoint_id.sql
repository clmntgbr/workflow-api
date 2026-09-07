-- +goose Up
ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_config_check;

DROP INDEX IF EXISTS idx_step_endpoint;

ALTER TABLE steps
    DROP COLUMN IF EXISTS endpoint_id;

ALTER TABLE steps
    ADD CONSTRAINT steps_type_config_check CHECK (
        (type = 'http' AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NULL)
        OR (type = 'delay' AND delay_duration_seconds IS NOT NULL AND delay_duration_seconds > 0 AND expression IS NULL)
        OR (type = 'condition' AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NOT NULL AND btrim(expression) <> '')
    );

ALTER TABLE step_runs
    DROP COLUMN IF EXISTS endpoint_id;

-- +goose Down
ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_config_check;

ALTER TABLE steps
    ADD COLUMN endpoint_id UUID NULL REFERENCES endpoints (id);

CREATE INDEX IF NOT EXISTS idx_step_endpoint ON steps (endpoint_id);

ALTER TABLE steps
    ADD CONSTRAINT steps_type_config_check CHECK (
        (type = 'http' AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NULL)
        OR (type = 'delay' AND endpoint_id IS NULL AND delay_duration_seconds IS NOT NULL AND delay_duration_seconds > 0 AND expression IS NULL)
        OR (type = 'condition' AND endpoint_id IS NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NOT NULL AND btrim(expression) <> '')
    );

ALTER TABLE step_runs
    ADD COLUMN endpoint_id UUID NULL REFERENCES endpoints (id);
