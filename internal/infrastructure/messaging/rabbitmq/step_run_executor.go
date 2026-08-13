package rabbitmq

import (
	"context"
	"encoding/json"
	"time"

	"go-api/internal/domain/port"

	"github.com/google/uuid"
)

type stepRunExecutePayload struct {
	EventID       string `json:"eventId"`
	StepRunID     string `json:"stepRunId"`
	StepID        string `json:"stepId"`
	WorkflowRunID string `json:"workflowRunId"`
}

type stepRunExecutor struct {
	publisher *Publisher
}

func NewStepRunExecutor(publisher *Publisher) port.StepRunExecutor {
	return &stepRunExecutor{publisher: publisher}
}

func (e *stepRunExecutor) Enqueue(ctx context.Context, job port.StepRunExecuteJob) error {
	payload, err := json.Marshal(stepRunExecutePayload{
		EventID:       uuid.New().String(),
		StepRunID:     job.StepRunID.String(),
		StepID:        job.StepID.String(),
		WorkflowRunID: job.WorkflowRunID.String(),
	})
	if err != nil {
		return err
	}

	return e.publisher.Publish(ctx, port.EventEnvelope{
		EventID:     uuid.New().String(),
		Type:        port.EventTypeStepRunExecute,
		AggregateID: job.StepRunID.String(),
		OccurredAt:  time.Now().UTC().Format(time.RFC3339Nano),
		Payload:     payload,
	})
}
