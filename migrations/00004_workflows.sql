-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workflows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'inactive',
    organization_id UUID NOT NULL REFERENCES organizations (id) ON DELETE CASCADE,
    schedule_interval_minutes INTEGER NOT NULL DEFAULT 0,
    concurrency INTEGER NOT NULL DEFAULT 1,
    notifications_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_success BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_failure BOOLEAN NOT NULL DEFAULT TRUE,
    notify_on_cancel BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT workflows_status_check CHECK (status IN ('active', 'inactive', 'deleted', 'canceled'))
);

CREATE INDEX IF NOT EXISTS idx_workflow_name ON workflows (name);
CREATE INDEX IF NOT EXISTS idx_workflow_status ON workflows (status);
CREATE INDEX IF NOT EXISTS idx_workflow_org ON workflows (organization_id);
CREATE INDEX IF NOT EXISTS idx_workflow_org_status ON workflows (organization_id, status);
CREATE INDEX IF NOT EXISTS idx_workflow_created ON workflows (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workflows;
-- +goose StatementEnd
