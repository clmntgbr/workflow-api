-- +goose Up
ALTER TABLE workflow_runs
    DROP CONSTRAINT IF EXISTS workflow_runs_status_check;
ALTER TABLE workflow_runs
    ADD CONSTRAINT workflow_runs_status_check
        CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled', 'skipped'));

ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_status_check;
ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_status_check
        CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled', 'skipped'));

-- +goose Down
ALTER TABLE workflow_runs
    DROP CONSTRAINT IF EXISTS workflow_runs_status_check;
ALTER TABLE workflow_runs
    ADD CONSTRAINT workflow_runs_status_check
        CHECK (status IN ('pending', 'running', 'success', 'failed', 'cancelled'));

ALTER TABLE step_runs
    DROP CONSTRAINT IF EXISTS step_runs_status_check;
ALTER TABLE step_runs
    ADD CONSTRAINT step_runs_status_check
        CHECK (status IN ('pending', 'running', 'success', 'failed', 'skipped'));
