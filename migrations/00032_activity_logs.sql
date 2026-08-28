-- +goose Up
CREATE TABLE activity_logs (
    id                  UUID PRIMARY KEY,
    project_id          UUID NOT NULL REFERENCES projects(id),
    action              TEXT NOT NULL,
    subject_type        TEXT NOT NULL,
    subject_id          UUID NOT NULL,
    workflow_id         UUID REFERENCES workflows(id) ON DELETE SET NULL,
    workflow_run_id     UUID REFERENCES workflow_runs(id) ON DELETE SET NULL,
    step_id             UUID REFERENCES steps(id) ON DELETE SET NULL,
    step_run_id         UUID REFERENCES step_runs(id) ON DELETE SET NULL,
    actor_type          TEXT,
    actor_user_id       UUID REFERENCES users(id) ON DELETE SET NULL,
    level               TEXT NOT NULL DEFAULT 'info',
    message             TEXT NOT NULL DEFAULT '',
    payload             JSONB NOT NULL DEFAULT '{}',
    source_event_id     UUID NOT NULL UNIQUE,
    source_event_type   TEXT NOT NULL,
    occurred_at         TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_activity_logs_project_occurred
    ON activity_logs (project_id, occurred_at DESC);

CREATE INDEX idx_activity_logs_workflow_occurred
    ON activity_logs (workflow_id, occurred_at DESC)
    WHERE workflow_id IS NOT NULL;

CREATE INDEX idx_activity_logs_workflow_run_occurred
    ON activity_logs (workflow_run_id, occurred_at DESC)
    WHERE workflow_run_id IS NOT NULL;

-- +goose Down
DROP TABLE IF EXISTS activity_logs;
