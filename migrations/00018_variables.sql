-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS variables (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    key VARCHAR(255) NOT NULL,
    description VARCHAR(255) NULL,
    path VARCHAR(255) NOT NULL,
    step_id UUID NOT NULL REFERENCES steps (id) ON DELETE CASCADE,
    workflow_id UUID NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_variables_workflow_key UNIQUE (workflow_id, key)
);

CREATE INDEX IF NOT EXISTS idx_variables_step ON variables (step_id);
CREATE INDEX IF NOT EXISTS idx_variables_workflow ON variables (workflow_id);

ALTER TABLE step_runs
    ADD COLUMN IF NOT EXISTS variable_extracts JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS extracted_variables JSONB NOT NULL DEFAULT '{}'::jsonb;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE step_runs
    DROP COLUMN IF EXISTS extracted_variables,
    DROP COLUMN IF EXISTS variable_extracts;

DROP TABLE IF EXISTS variables;
-- +goose StatementEnd
