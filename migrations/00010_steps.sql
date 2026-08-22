-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workflow_id UUID NOT NULL REFERENCES workflows (id) ON DELETE CASCADE,
    endpoint_id UUID NOT NULL REFERENCES endpoints (id),
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
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
    position_x INTEGER NOT NULL DEFAULT 0,
    position_y INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT steps_method_check CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS')),
    CONSTRAINT steps_status_check CHECK (status IN ('active', 'deleted'))
);

CREATE INDEX IF NOT EXISTS idx_step_workflow ON steps (workflow_id);
CREATE INDEX IF NOT EXISTS idx_step_endpoint ON steps (endpoint_id);
CREATE INDEX IF NOT EXISTS idx_step_org ON steps (project_id);
CREATE INDEX IF NOT EXISTS idx_step_execution_order ON steps (execution_order);
CREATE INDEX IF NOT EXISTS idx_step_created ON steps (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS steps;
-- +goose StatementEnd
