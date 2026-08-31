package port

import (
	"context"

	"github.com/google/uuid"
)

const EventTypeStepRunExecute = "step_run.execute"

type StepRunExecuteJob struct {
	StepRunID     uuid.UUID
	StepID        uuid.UUID
	WorkflowRunID uuid.UUID
	Priority      uint8
}

type StepRunExecutor interface {
	Enqueue(ctx context.Context, job StepRunExecuteJob) error
}
