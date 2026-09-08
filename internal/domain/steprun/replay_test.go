package steprun

import (
	"testing"
	"time"

	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

func TestStepRunView_ReplayOnto(t *testing.T) {
	started := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	finished := started.Add(2 * time.Second)
	sourceRunID := uuid.New()
	stepID := uuid.New()
	newRunID := uuid.New()

	view := StepRunView{
		ID:                 uuid.New(),
		WorkflowRunID:      sourceRunID,
		StepID:             stepID,
		StepType:           domainstep.TypeHTTP,
		Name:               "Login",
		Status:             StatusSuccess,
		ExtractedVariables: map[string]any{"token": "abc"},
		StartedAt:          &started,
		FinishedAt:         &finished,
	}

	cloned := view.ReplayOnto(newRunID)
	if cloned.ID == view.ID {
		t.Fatal("replayed step run must get a new id")
	}
	if cloned.WorkflowRunID != newRunID {
		t.Fatalf("workflow run id: got %s", cloned.WorkflowRunID)
	}
	if cloned.StepID != stepID {
		t.Fatalf("step id: got %s", cloned.StepID)
	}
	if cloned.Status != StatusSuccess {
		t.Fatalf("status: got %s", cloned.Status)
	}
	if cloned.ExtractedVariables["token"] != "abc" {
		t.Fatalf("extracted variables: got %v", cloned.ExtractedVariables)
	}
	if len(cloned.PullEvents()) != 0 {
		t.Fatal("replayed step run must not emit events")
	}
}
