-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS run_exports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows (id),
    project_id UUID NOT NULL REFERENCES projects (id),
    requested_by_user_id UUID NOT NULL REFERENCES users (id),
    date_range_from TIMESTAMPTZ NOT NULL,
    date_range_to TIMESTAMPTZ NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT run_exports_status_check CHECK (status IN ('pending', 'processing', 'ready', 'failed')),
    CONSTRAINT run_exports_date_range_check CHECK (date_range_from <= date_range_to)
);

CREATE INDEX IF NOT EXISTS idx_run_exports_workflow ON run_exports (workflow_id);
CREATE INDEX IF NOT EXISTS idx_run_exports_project ON run_exports (project_id);
CREATE INDEX IF NOT EXISTS idx_run_exports_requested_by ON run_exports (requested_by_user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS run_exports;
-- +goose StatementEnd
