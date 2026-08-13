package steprun

import (
	"context"
	"encoding/json"
	"log"

	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
)

type ExecuteJob struct {
	EventID       string `json:"eventId"`
	StepRunID     string `json:"stepRunId"`
	StepID        string `json:"stepId"`
	WorkflowRunID string `json:"workflowRunId"`
}

type ExecuteHandler struct{}

func NewExecuteHandler() *ExecuteHandler {
	return &ExecuteHandler{}
}

func (h *ExecuteHandler) Handle(_ context.Context, payload []byte) error {
	var job ExecuteJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return messaging.NonRetryable(err)
	}
	if job.StepRunID == "" || job.StepID == "" || job.WorkflowRunID == "" {
		return messaging.NonRetryable(errInvalidExecuteJob)
	}

	log.Printf(
		"executor received %s stepRunId=%s stepId=%s workflowRunId=%s",
		port.EventTypeStepRunExecute,
		job.StepRunID,
		job.StepID,
		job.WorkflowRunID,
	)
	return nil
}

type invalidExecuteJobError struct{}

func (invalidExecuteJobError) Error() string { return "invalid step run execute job" }

var errInvalidExecuteJob error = invalidExecuteJobError{}
