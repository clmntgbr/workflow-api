-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX IF NOT EXISTS uq_workflow_runs_in_progress
    ON workflow_runs (workflow_id)
    WHERE status IN ('pending', 'running');
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS uq_workflow_runs_in_progress;
-- +goose StatementEnd
