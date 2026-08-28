-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS assertions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    description VARCHAR(255) NULL,
    source VARCHAR(50) NOT NULL,
    path VARCHAR(255) NULL,
    operator VARCHAR(50) NOT NULL,
    expected_value TEXT NULL,
    step_id UUID NOT NULL REFERENCES steps (id) ON DELETE CASCADE,
    workflow_id UUID NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_assertions_source CHECK (source IN ('status', 'header', 'body')),
    CONSTRAINT chk_assertions_operator CHECK (operator IN (
        'equals',
        'not_equals',
        'not_null',
        'is_null',
        'contains',
        'greater_than',
        'less_than',
        'matches_regex',
        'is_string',
        'is_number',
        'is_boolean',
        'is_array',
        'is_object'
    ))
);

CREATE INDEX IF NOT EXISTS idx_assertions_step ON assertions (step_id);
CREATE INDEX IF NOT EXISTS idx_assertions_workflow ON assertions (workflow_id);

ALTER TABLE step_runs
    ADD COLUMN IF NOT EXISTS assertions JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS assertions_result JSONB NOT NULL DEFAULT '[]'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE step_runs
    DROP COLUMN IF EXISTS assertions_result,
    DROP COLUMN IF EXISTS assertions;

DROP TABLE IF EXISTS assertions;
-- +goose StatementEnd
