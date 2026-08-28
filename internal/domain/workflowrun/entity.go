package workflowrun

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type WorkflowRun struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	Status            Status
	TriggeredBy       TriggeredBy
	TriggeredByUserID *uuid.UUID
	Context           map[string]any
	StartedAt         *time.Time
	FinishedAt        *time.Time
	Error             string
	CreatedAt         time.Time
	UpdatedAt         time.Time

	events []event.DomainEvent
}

type NewWorkflowRunParams struct {
	WorkflowID        uuid.UUID
	TriggeredBy       TriggeredBy
	TriggeredByUserID *uuid.UUID
	Context           map[string]any
}

func NewWorkflowRun(p NewWorkflowRunParams) *WorkflowRun {
	now := time.Now().UTC()
	ctx := p.Context
	if ctx == nil {
		ctx = map[string]any{}
	}

	run := &WorkflowRun{
		ID:                uuid.New(),
		WorkflowID:        p.WorkflowID,
		Status:            StatusPending,
		TriggeredBy:       p.TriggeredBy,
		TriggeredByUserID: p.TriggeredByUserID,
		Context:           ctx,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	run.recordEvent(run.startedEvent(now))
	return run
}

func (r *WorkflowRun) MarkRunning() error {
	if r.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if r.Status != StatusPending {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	r.Status = StatusRunning
	r.StartedAt = &now
	r.UpdatedAt = now
	return nil
}

func (r *WorkflowRun) MarkSucceeded() error {
	if r.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if r.Status != StatusPending && r.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	if r.StartedAt == nil {
		r.StartedAt = &now
	}
	r.Status = StatusSuccess
	r.FinishedAt = &now
	r.Error = ""
	r.UpdatedAt = now
	r.recordEvent(r.succeededEvent(now))
	r.recordEvent(r.finishedEvent(now))
	return nil
}

func (r *WorkflowRun) MarkFailed(errMsg string) error {
	if r.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if r.Status != StatusPending && r.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	if r.StartedAt == nil {
		r.StartedAt = &now
	}
	r.Status = StatusFailed
	r.FinishedAt = &now
	r.Error = errMsg
	r.UpdatedAt = now
	r.recordEvent(r.failedEvent(now))
	r.recordEvent(r.finishedEvent(now))
	return nil
}

func (r *WorkflowRun) MarkCancelled() error {
	if r.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if r.Status != StatusPending && r.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	r.Status = StatusCancelled
	r.FinishedAt = &now
	r.UpdatedAt = now
	r.recordEvent(r.cancelledEvent(now))
	r.recordEvent(r.finishedEvent(now))
	return nil
}

func (r *WorkflowRun) MergeContext(vars map[string]any) {
	if len(vars) == 0 {
		return
	}
	if r.Context == nil {
		r.Context = map[string]any{}
	}
	for key, value := range vars {
		r.Context[key] = value
	}
	r.UpdatedAt = time.Now().UTC()
}

func (r *WorkflowRun) PullEvents() []event.DomainEvent {
	events := r.events
	r.events = nil
	return events
}

func (r *WorkflowRun) recordEvent(e event.DomainEvent) {
	r.events = append(r.events, e)
}

func (r *WorkflowRun) triggeredByUserIDString() *string {
	if r.TriggeredByUserID == nil {
		return nil
	}
	value := r.TriggeredByUserID.String()
	return &value
}

func (r *WorkflowRun) startedEvent(at time.Time) WorkflowRunStarted {
	return WorkflowRunStarted{
		ID:                uuid.New().String(),
		WorkflowRunID:     r.ID.String(),
		WorkflowID:        r.WorkflowID.String(),
		Status:            string(r.Status),
		TriggeredBy:       string(r.TriggeredBy),
		TriggeredByUserID: r.triggeredByUserIDString(),
		Timestamp:         at,
	}
}

func (r *WorkflowRun) succeededEvent(at time.Time) WorkflowRunSucceeded {
	return WorkflowRunSucceeded{
		ID:            uuid.New().String(),
		WorkflowRunID: r.ID.String(),
		WorkflowID:    r.WorkflowID.String(),
		Status:        string(r.Status),
		Timestamp:     at,
	}
}

func (r *WorkflowRun) failedEvent(at time.Time) WorkflowRunFailed {
	return WorkflowRunFailed{
		ID:            uuid.New().String(),
		WorkflowRunID: r.ID.String(),
		WorkflowID:    r.WorkflowID.String(),
		Status:        string(r.Status),
		Error:         r.Error,
		Timestamp:     at,
	}
}

func (r *WorkflowRun) cancelledEvent(at time.Time) WorkflowRunCancelled {
	return WorkflowRunCancelled{
		ID:            uuid.New().String(),
		WorkflowRunID: r.ID.String(),
		WorkflowID:    r.WorkflowID.String(),
		Status:        string(r.Status),
		Timestamp:     at,
	}
}

func (r *WorkflowRun) finishedEvent(at time.Time) WorkflowRunFinished {
	return WorkflowRunFinished{
		ID:            uuid.New().String(),
		WorkflowRunID: r.ID.String(),
		WorkflowID:    r.WorkflowID.String(),
		FinishType:    finishTypeFromStatus(r.Status),
		Status:        string(r.Status),
		Error:         r.Error,
		Timestamp:     at,
	}
}

func finishTypeFromStatus(status Status) FinishType {
	switch status {
	case StatusSuccess:
		return FinishTypeSuccess
	case StatusFailed:
		return FinishTypeFailed
	case StatusCancelled:
		return FinishTypeCancelled
	default:
		return FinishType(status)
	}
}
