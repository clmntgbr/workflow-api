package steprun

import (
	"context"
	"testing"
	"time"

	"go-api/internal/domain/event"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

func TestFailStaleStepRunsHandler_FailsStaleHTTPAndCancelsWaitingDelay(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	pendingMaxAge := 30 * time.Minute
	grace := 5 * time.Minute

	workflowRunID := uuid.New()
	run := runningWorkflowRun(workflowRunID)

	staleHTTP := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeHTTP,
		Status:        domainsteprun.StatusPending,
		CreatedAt:     now.Add(-31 * time.Minute),
		Timeout:       7_200_000,
	}
	waitingDelay := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeDelay,
		Status:        domainsteprun.StatusWaiting,
		CreatedAt:     now.Add(-2 * time.Hour),
	}

	stepRepo := newMemStepRunRepo(staleHTTP, waitingDelay)
	runRepo := newMemWorkflowRunRepo(run)
	outbox := &memOutbox{}

	handler := NewFailStaleStepRunsHandler(stepRepo, runRepo, outbox, pendingMaxAge, grace, 1)
	failed, err := handler.Handle(context.Background(), now, 100)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if failed != 1 {
		t.Fatalf("failed runs = %d, want 1", failed)
	}

	if run.Status != domainworkflowrun.StatusFailed {
		t.Fatalf("workflow run status = %s, want failed", run.Status)
	}
	if run.Error != staleWorkflowRunError {
		t.Fatalf("workflow run error = %q", run.Error)
	}
	if staleHTTP.Status != domainsteprun.StatusFailed {
		t.Fatalf("stale HTTP status = %s, want failed", staleHTTP.Status)
	}
	if staleHTTP.Error != staleStepRunError {
		t.Fatalf("stale HTTP error = %q", staleHTTP.Error)
	}
	if waitingDelay.Status != domainsteprun.StatusCancelled {
		t.Fatalf("waiting delay status = %s, want cancelled", waitingDelay.Status)
	}
	if len(outbox.events) == 0 {
		t.Fatal("expected outbox events")
	}
}

func TestFailStaleStepRunsHandler_IgnoresLongRunningHTTPWithinTimeout(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	started := now.Add(-40 * time.Minute)

	workflowRunID := uuid.New()
	run := runningWorkflowRun(workflowRunID)
	httpRun := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeHTTP,
		Status:        domainsteprun.StatusRunning,
		Timeout:       3_600_000,
		StartedAt:     &started,
		CreatedAt:     started,
	}

	handler := NewFailStaleStepRunsHandler(
		newMemStepRunRepo(httpRun),
		newMemWorkflowRunRepo(run),
		&memOutbox{},
		30*time.Minute,
		5*time.Minute,
		1,
	)
	failed, err := handler.Handle(context.Background(), now, 100)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if failed != 0 {
		t.Fatalf("failed runs = %d, want 0", failed)
	}
	if run.Status != domainworkflowrun.StatusRunning {
		t.Fatalf("workflow run status = %s, want running", run.Status)
	}
	if httpRun.Status != domainsteprun.StatusRunning {
		t.Fatalf("HTTP status = %s, want running", httpRun.Status)
	}
}

func TestFailStaleStepRunsHandler_IgnoresWaitingDelayOnly(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	workflowRunID := uuid.New()
	run := runningWorkflowRun(workflowRunID)
	delay := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeDelay,
		Status:        domainsteprun.StatusWaiting,
		CreatedAt:     now.Add(-3 * time.Hour),
	}

	handler := NewFailStaleStepRunsHandler(
		newMemStepRunRepo(delay),
		newMemWorkflowRunRepo(run),
		&memOutbox{},
		30*time.Minute,
		5*time.Minute,
		1,
	)
	failed, err := handler.Handle(context.Background(), now, 100)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if failed != 0 {
		t.Fatalf("failed runs = %d, want 0", failed)
	}
	if run.Status != domainworkflowrun.StatusRunning {
		t.Fatalf("workflow run should stay running during a long delay")
	}
}

func TestFailStaleStepRunsHandler_DeduplicatesByWorkflowRun(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	workflowRunID := uuid.New()
	run := runningWorkflowRun(workflowRunID)

	first := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeHTTP,
		Status:        domainsteprun.StatusPending,
		CreatedAt:     now.Add(-40 * time.Minute),
	}
	second := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeCondition,
		Status:        domainsteprun.StatusPending,
		CreatedAt:     now.Add(-35 * time.Minute),
	}

	handler := NewFailStaleStepRunsHandler(
		newMemStepRunRepo(first, second),
		newMemWorkflowRunRepo(run),
		&memOutbox{},
		30*time.Minute,
		5*time.Minute,
		1,
	)
	failed, err := handler.Handle(context.Background(), now, 100)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if failed != 1 {
		t.Fatalf("failed runs = %d, want 1", failed)
	}
	if first.Status != domainsteprun.StatusFailed || second.Status != domainsteprun.StatusFailed {
		t.Fatalf("both stale steps should be failed, got %s and %s", first.Status, second.Status)
	}
}

func TestFailStaleStepRunsHandler_FailsRunningHTTPPastTimeout(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	started := now.Add(-15 * time.Minute)
	workflowRunID := uuid.New()
	run := runningWorkflowRun(workflowRunID)
	httpRun := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeHTTP,
		Status:        domainsteprun.StatusRunning,
		Timeout:       30_000,
		StartedAt:     &started,
		CreatedAt:     started,
	}

	handler := NewFailStaleStepRunsHandler(
		newMemStepRunRepo(httpRun),
		newMemWorkflowRunRepo(run),
		&memOutbox{},
		30*time.Minute,
		5*time.Minute,
		1,
	)
	failed, err := handler.Handle(context.Background(), now, 100)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if failed != 1 {
		t.Fatalf("failed runs = %d, want 1", failed)
	}
	if run.Status != domainworkflowrun.StatusFailed {
		t.Fatalf("workflow run status = %s, want failed", run.Status)
	}
	if httpRun.Status != domainsteprun.StatusFailed {
		t.Fatalf("HTTP status = %s, want failed", httpRun.Status)
	}
}

func TestFailStaleStepRunsHandler_SkipsAlreadyTerminalWorkflowRun(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	workflowRunID := uuid.New()
	run := runningWorkflowRun(workflowRunID)
	if err := run.MarkSucceeded(); err != nil {
		t.Fatal(err)
	}

	stale := &domainsteprun.StepRun{
		ID:            uuid.New(),
		WorkflowRunID: workflowRunID,
		StepType:      domainstep.TypeHTTP,
		Status:        domainsteprun.StatusPending,
		CreatedAt:     now.Add(-40 * time.Minute),
	}

	handler := NewFailStaleStepRunsHandler(
		newMemStepRunRepo(stale),
		newMemWorkflowRunRepo(run),
		&memOutbox{},
		30*time.Minute,
		5*time.Minute,
		1,
	)
	failed, err := handler.Handle(context.Background(), now, 100)
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if failed != 0 {
		t.Fatalf("failed runs = %d, want 0", failed)
	}
	if stale.Status != domainsteprun.StatusPending {
		t.Fatalf("step should be left untouched when the run is already terminal")
	}
}

func runningWorkflowRun(id uuid.UUID) *domainworkflowrun.WorkflowRun {
	run := domainworkflowrun.NewWorkflowRun(domainworkflowrun.NewWorkflowRunParams{
		WorkflowID: uuid.New(),
	})
	run.ID = id
	_ = run.MarkRunning()
	_ = run.PullEvents()
	return run
}

type memStepRunRepo struct {
	byID map[uuid.UUID]*domainsteprun.StepRun
}

func newMemStepRunRepo(runs ...*domainsteprun.StepRun) *memStepRunRepo {
	byID := make(map[uuid.UUID]*domainsteprun.StepRun, len(runs))
	for _, run := range runs {
		byID[run.ID] = run
	}
	return &memStepRunRepo{byID: byID}
}

func (m *memStepRunRepo) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *memStepRunRepo) Save(context.Context, *domainsteprun.StepRun) error {
	return nil
}

func (m *memStepRunRepo) Update(_ context.Context, run *domainsteprun.StepRun) error {
	m.byID[run.ID] = run
	return nil
}

func (m *memStepRunRepo) GetByID(_ context.Context, id uuid.UUID) (*domainsteprun.StepRun, error) {
	return m.byID[id], nil
}

func (m *memStepRunRepo) FindByWorkflowRunID(_ context.Context, workflowRunID uuid.UUID) ([]*domainsteprun.StepRun, error) {
	out := make([]*domainsteprun.StepRun, 0)
	for _, run := range m.byID {
		if run.WorkflowRunID == workflowRunID {
			out = append(out, run)
		}
	}
	return out, nil
}

func (m *memStepRunRepo) ClaimDueWaiting(context.Context, time.Time, int) ([]*domainsteprun.StepRun, error) {
	return nil, nil
}

func (m *memStepRunRepo) FindActiveNonDelay(
	_ context.Context,
	now time.Time,
	pendingMaxAge, grace time.Duration,
	limit int,
) ([]*domainsteprun.StepRun, error) {
	out := make([]*domainsteprun.StepRun, 0)
	for _, run := range m.byID {
		if run.IsStale(now, pendingMaxAge, grace) {
			out = append(out, run)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

type memWorkflowRunRepo struct {
	byID map[uuid.UUID]*domainworkflowrun.WorkflowRun
}

func newMemWorkflowRunRepo(runs ...*domainworkflowrun.WorkflowRun) *memWorkflowRunRepo {
	byID := make(map[uuid.UUID]*domainworkflowrun.WorkflowRun, len(runs))
	for _, run := range runs {
		byID[run.ID] = run
	}
	return &memWorkflowRunRepo{byID: byID}
}

func (m *memWorkflowRunRepo) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func (m *memWorkflowRunRepo) Save(context.Context, *domainworkflowrun.WorkflowRun) error {
	return nil
}

func (m *memWorkflowRunRepo) Update(_ context.Context, run *domainworkflowrun.WorkflowRun) error {
	m.byID[run.ID] = run
	return nil
}

func (m *memWorkflowRunRepo) GetByID(_ context.Context, id uuid.UUID) (*domainworkflowrun.WorkflowRun, error) {
	return m.byID[id], nil
}

func (m *memWorkflowRunRepo) HasInProgress(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}

func (m *memWorkflowRunRepo) FindInProgressByWorkflowID(
	context.Context,
	uuid.UUID,
) (*domainworkflowrun.WorkflowRun, error) {
	return nil, nil
}

type memOutbox struct {
	events []event.DomainEvent
}

func (m *memOutbox) StoreEvents(_ context.Context, events []event.DomainEvent) error {
	m.events = append(m.events, events...)
	return nil
}

func (m *memOutbox) FetchUnpublished(context.Context, int) ([]port.OutboxMessage, error) {
	return nil, nil
}

func (m *memOutbox) MarkPublished(context.Context, []uuid.UUID) error {
	return nil
}
