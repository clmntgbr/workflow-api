package workflowrun_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	eventworkflowrun "go-api/internal/application/event/workflowrun"
	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type mockStepRunExecutor struct {
	called bool
	job    port.StepRunExecuteJob
	err    error
}

func (m *mockStepRunExecutor) Enqueue(_ context.Context, job port.StepRunExecuteJob) error {
	m.called = true
	m.job = job
	return m.err
}

type mockExecutorPriorityResolver struct {
	called    bool
	projectID uuid.UUID
	priority  uint8
	err       error
}

func (m *mockExecutorPriorityResolver) ExecutorPriorityForProject(
	_ context.Context,
	projectID uuid.UUID,
) (uint8, error) {
	m.called = true
	m.projectID = projectID
	if m.err != nil {
		return 0, m.err
	}
	return m.priority, nil
}

func TestEnqueueStepRunHandler_UsesExecutorPriority(t *testing.T) {
	projectID := uuid.MustParse("01960000-0000-7000-8000-00000000000a")
	stepRunID := uuid.MustParse("01960000-0000-7000-8000-00000000000b")
	stepID := uuid.MustParse("01960000-0000-7000-8000-00000000000c")
	workflowRunID := uuid.MustParse("01960000-0000-7000-8000-00000000000d")

	executor := &mockStepRunExecutor{}
	resolver := &mockExecutorPriorityResolver{priority: 5}
	handler := eventworkflowrun.NewEnqueueStepRunHandler(executor, resolver)

	payload, err := json.Marshal(domainsteprun.StepRunQueued{
		ID:            uuid.New().String(),
		StepRunID:     stepRunID.String(),
		WorkflowRunID: workflowRunID.String(),
		StepID:        stepID.String(),
		ProjectID:     projectID.String(),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if err := handler.Handle(context.Background(), payload); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if !executor.called {
		t.Fatal("expected executor to be called")
	}
	if !resolver.called {
		t.Fatal("expected priority resolver to be called")
	}
	if resolver.projectID != projectID {
		t.Fatalf("project id: got %s", resolver.projectID)
	}
	if executor.job.Priority != 5 {
		t.Fatalf("priority: got %d want 5", executor.job.Priority)
	}
}

func TestEnqueueStepRunHandler_FallsBackToZeroPriorityOnResolverError(t *testing.T) {
	executor := &mockStepRunExecutor{}
	resolver := &mockExecutorPriorityResolver{err: errors.New("subscription unavailable")}
	handler := eventworkflowrun.NewEnqueueStepRunHandler(executor, resolver)

	payload, err := json.Marshal(domainsteprun.StepRunQueued{
		ID:            uuid.New().String(),
		StepRunID:     uuid.New().String(),
		WorkflowRunID: uuid.New().String(),
		StepID:        uuid.New().String(),
		ProjectID:     uuid.New().String(),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if err := handler.Handle(context.Background(), payload); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if executor.job.Priority != 0 {
		t.Fatalf("priority: got %d want 0", executor.job.Priority)
	}
}

func TestEnqueueStepRunHandler_MissingProjectIDUsesZeroPriority(t *testing.T) {
	executor := &mockStepRunExecutor{}
	resolver := &mockExecutorPriorityResolver{priority: 8}
	handler := eventworkflowrun.NewEnqueueStepRunHandler(executor, resolver)

	payload, err := json.Marshal(domainsteprun.StepRunQueued{
		ID:            uuid.New().String(),
		StepRunID:     uuid.New().String(),
		WorkflowRunID: uuid.New().String(),
		StepID:        uuid.New().String(),
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	if err := handler.Handle(context.Background(), payload); err != nil {
		t.Fatalf("handle: %v", err)
	}
	if resolver.called {
		t.Fatal("resolver must not be called without project id")
	}
	if executor.job.Priority != 0 {
		t.Fatalf("priority: got %d want 0", executor.job.Priority)
	}
}
