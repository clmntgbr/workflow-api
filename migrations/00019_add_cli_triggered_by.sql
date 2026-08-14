-- +goose Up
-- +goose StatementBegin
ALTER TABLE workflow_runs
DROP CONSTRAINT workflow_runs_triggered_by_check,
ADD CONSTRAINT workflow_runs_triggered_by_check CHECK (triggered_by IN ('user', 'schedule', 'webhook', 'api', 'cli'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE workflow_runs
DROP CONSTRAINT workflow_runs_triggered_by_check,
ADD CONSTRAINT workflow_runs_triggered_by_check CHECK (triggered_by IN ('user', 'schedule', 'webhook', 'api'));
-- +goose StatementEnd
