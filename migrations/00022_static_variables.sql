-- +goose Up
-- +goose StatementBegin
ALTER TABLE variables
    ALTER COLUMN step_id DROP NOT NULL,
    ALTER COLUMN path DROP NOT NULL;

ALTER TABLE variables
    ADD COLUMN IF NOT EXISTS kind VARCHAR(32) NOT NULL DEFAULT 'extracted',
    ADD COLUMN IF NOT EXISTS value JSONB;

UPDATE variables
SET kind = 'extracted'
WHERE kind IS NULL OR kind = '';

ALTER TABLE variables
    DROP CONSTRAINT IF EXISTS chk_variables_kind_shape;

ALTER TABLE variables
    ADD CONSTRAINT chk_variables_kind_shape CHECK (
        (
            kind = 'extracted'
            AND step_id IS NOT NULL
            AND path IS NOT NULL
            AND length(trim(path)) > 0
            AND value IS NULL
        )
        OR
        (
            kind = 'static'
            AND step_id IS NULL
            AND (path IS NULL OR length(trim(path)) = 0)
            AND value IS NOT NULL
        )
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE variables DROP CONSTRAINT IF EXISTS chk_variables_kind_shape;

DELETE FROM variables WHERE kind = 'static';

ALTER TABLE variables DROP COLUMN IF EXISTS value;
ALTER TABLE variables DROP COLUMN IF EXISTS kind;

UPDATE variables SET path = '' WHERE path IS NULL;

ALTER TABLE variables
    ALTER COLUMN path SET NOT NULL;

ALTER TABLE variables
    ALTER COLUMN step_id SET NOT NULL;
-- +goose StatementEnd
