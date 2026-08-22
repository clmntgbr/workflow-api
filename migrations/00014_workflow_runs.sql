-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workflow_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows (id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    triggered_by VARCHAR(20) NOT NULL,
    triggered_by_user_id UUID NULL REFERENCES users (id) ON DELETE SET NULL,
    context JSONB NOT NULL DEFAULT '{}'::jsonb,
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT workflow_runs_status_check CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled')),
    CONSTRAINT workflow_runs_triggered_by_check CHECK (triggered_by IN ('user', 'schedule', 'webhook', 'api'))
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow ON workflow_runs (workflow_id);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_status ON workflow_runs (status);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_created ON workflow_runs (created_at);
CREATE INDEX IF NOT EXISTS idx_workflow_runs_workflow_created ON workflow_runs (workflow_id, created_at);

CREATE TABLE IF NOT EXISTS step_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_run_id UUID NOT NULL REFERENCES workflow_runs (id) ON DELETE CASCADE,
    step_id UUID NOT NULL REFERENCES steps (id),
    workflow_id UUID NOT NULL REFERENCES workflows (id),
    endpoint_id UUID NOT NULL REFERENCES endpoints (id),
    project_id UUID NOT NULL REFERENCES projects (id),
    name TEXT NOT NULL,
    description TEXT NULL,
    url TEXT NOT NULL,
    method VARCHAR(10) NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    query_params JSONB NOT NULL DEFAULT '{}'::jsonb,
    body JSONB NOT NULL DEFAULT '{}'::jsonb,
    timeout_ms INTEGER NOT NULL DEFAULT 30000,
    retry_on_failure BOOLEAN NOT NULL DEFAULT FALSE,
    retry_count INTEGER NOT NULL DEFAULT 0,
    retry_delay_ms INTEGER NOT NULL DEFAULT 1000,
    step_index VARCHAR(255) NOT NULL,
    execution_order INTEGER NOT NULL DEFAULT 0,
    tree_index INTEGER NOT NULL DEFAULT 0,
    position_x DOUBLE PRECISION NOT NULL DEFAULT 0,
    position_y DOUBLE PRECISION NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    attempt INTEGER NOT NULL DEFAULT 0,
    response_snapshot JSONB NULL,
    started_at TIMESTAMPTZ NULL,
    finished_at TIMESTAMPTZ NULL,
    error TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT step_runs_method_check CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS')),
    CONSTRAINT step_runs_status_check CHECK (status IN ('pending', 'running', 'success', 'failed', 'skipped')),
    CONSTRAINT uq_step_runs_workflow_run_step UNIQUE (workflow_run_id, step_id)
);

CREATE INDEX IF NOT EXISTS idx_step_runs_workflow_run ON step_runs (workflow_run_id);
CREATE INDEX IF NOT EXISTS idx_step_runs_step ON step_runs (step_id);
CREATE INDEX IF NOT EXISTS idx_step_runs_workflow ON step_runs (workflow_id);
CREATE INDEX IF NOT EXISTS idx_step_runs_execution_order ON step_runs (execution_order);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS step_runs;
DROP TABLE IF EXISTS workflow_runs;
-- +goose StatementEnd
