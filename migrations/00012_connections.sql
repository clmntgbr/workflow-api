-- +goose Up
CREATE TABLE connections (
    id UUID PRIMARY KEY,
    workflow_id UUID NOT NULL REFERENCES workflows(id),
    organization_id UUID NOT NULL REFERENCES organizations(id),
    source_step_id UUID NOT NULL REFERENCES steps(id),
    target_step_id UUID NOT NULL REFERENCES steps(id),

    CONSTRAINT uq_connection_source_target UNIQUE (workflow_id, source_step_id, target_step_id)
);

CREATE INDEX idx_connections_workflow_id ON connections(workflow_id);

-- +goose Down
DROP TABLE IF EXISTS connections;
