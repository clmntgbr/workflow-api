-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    description TEXT NULL,
    url TEXT NOT NULL,
    method VARCHAR(10) NOT NULL,
    headers JSONB NOT NULL DEFAULT '{}'::jsonb,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    project_id UUID NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT endpoints_method_check CHECK (method IN ('GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS')),
    CONSTRAINT endpoints_status_check CHECK (status IN ('active', 'inactive', 'deleted'))
);

CREATE INDEX IF NOT EXISTS idx_endpoint_name ON endpoints (name);
CREATE INDEX IF NOT EXISTS idx_endpoint_status ON endpoints (status);
CREATE INDEX IF NOT EXISTS idx_endpoint_org ON endpoints (project_id);
CREATE INDEX IF NOT EXISTS idx_endpoint_org_status ON endpoints (project_id, status);
CREATE INDEX IF NOT EXISTS idx_endpoint_created ON endpoints (created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS endpoints;
-- +goose StatementEnd
