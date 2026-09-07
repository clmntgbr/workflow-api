-- +goose Up
ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_config_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_type_config_check CHECK (
        (type = 'http' AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NULL)
        OR (type = 'delay' AND endpoint_id IS NULL AND delay_duration_seconds IS NOT NULL AND delay_duration_seconds > 0 AND expression IS NULL)
        OR (type = 'condition' AND endpoint_id IS NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NOT NULL AND btrim(expression) <> '')
    );

-- +goose Down
ALTER TABLE steps
    DROP CONSTRAINT IF EXISTS steps_type_config_check;

ALTER TABLE steps
    ADD CONSTRAINT steps_type_config_check CHECK (
        (type = 'http' AND endpoint_id IS NOT NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NULL)
        OR (type = 'delay' AND endpoint_id IS NULL AND delay_duration_seconds IS NOT NULL AND delay_duration_seconds > 0 AND expression IS NULL)
        OR (type = 'condition' AND endpoint_id IS NULL AND (delay_duration_seconds IS NULL OR delay_duration_seconds = 0) AND expression IS NOT NULL AND btrim(expression) <> '')
    );
