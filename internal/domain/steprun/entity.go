package steprun

import (
	"time"

	"go-api/internal/domain/event"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type StepRun struct {
	ID            uuid.UUID
	WorkflowRunID uuid.UUID
	StepID        uuid.UUID

	WorkflowID     uuid.UUID
	EndpointID     uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         string
	Headers        map[string]string
	Query          map[string]string
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Index          string
	ExecutionOrder int
	TreeIndex      int
	Position       domainstep.Position

	Status  Status
	Attempt int

	ResponseSnapshot *ResponseSnapshot

	StartedAt  *time.Time
	FinishedAt *time.Time
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time

	events []event.DomainEvent
}

type NewStepRunParams struct {
	WorkflowRunID  uuid.UUID
	StepID         uuid.UUID
	WorkflowID     uuid.UUID
	EndpointID     uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         string
	Headers        map[string]string
	Query          map[string]string
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Index          string
	ExecutionOrder int
	TreeIndex      int
	Position       domainstep.Position
}

func NewStepRun(p NewStepRunParams) *StepRun {
	now := time.Now().UTC()
	return &StepRun{
		ID:             uuid.New(),
		WorkflowRunID:  p.WorkflowRunID,
		StepID:         p.StepID,
		WorkflowID:     p.WorkflowID,
		EndpointID:     p.EndpointID,
		OrganizationID: p.OrganizationID,
		Name:           p.Name,
		Description:    p.Description,
		URL:            p.URL,
		Method:         p.Method,
		Headers:        normalizeStringMap(p.Headers),
		Query:          normalizeStringMap(p.Query),
		Body:           normalizeAnyMap(p.Body),
		Timeout:        p.Timeout,
		RetryOnFailure: p.RetryOnFailure,
		RetryCount:     p.RetryCount,
		RetryDelay:     p.RetryDelay,
		Index:          p.Index,
		ExecutionOrder: p.ExecutionOrder,
		TreeIndex:      p.TreeIndex,
		Position:       p.Position,
		Status:         StatusPending,
		Attempt:        0,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// Queue records that this pending StepRun is ready for the executor.
func (s *StepRun) Queue() {
	s.recordEvent(StepRunQueued{
		ID:            uuid.New().String(),
		StepRunID:     s.ID.String(),
		WorkflowRunID: s.WorkflowRunID.String(),
		StepID:        s.StepID.String(),
		Timestamp:     time.Now().UTC(),
	})
}

func (s *StepRun) MarkStarted() error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusPending {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	s.Status = StatusRunning
	if s.Attempt == 0 {
		s.Attempt = 1
	}
	s.StartedAt = &now
	s.UpdatedAt = now
	s.recordEvent(s.startedEvent(now))
	return nil
}

// IncrementAttempt records another try on the same row. Status stays running.
func (s *StepRun) IncrementAttempt() error {
	if s.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}
	s.Attempt++
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *StepRun) MarkSucceeded(response ResponseSnapshot) error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	normalized := response.normalized()
	s.Status = StatusSuccess
	s.ResponseSnapshot = &normalized
	s.Error = ""
	s.FinishedAt = &now
	s.UpdatedAt = now
	s.recordEvent(s.succeededEvent(now))
	return nil
}

func (s *StepRun) MarkFailed(errMsg string, response *ResponseSnapshot) error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusPending && s.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	if s.StartedAt == nil {
		s.StartedAt = &now
	}
	if response != nil {
		normalized := response.normalized()
		s.ResponseSnapshot = &normalized
	}
	s.Status = StatusFailed
	s.Error = errMsg
	s.FinishedAt = &now
	s.UpdatedAt = now
	s.recordEvent(s.failedEvent(now))
	return nil
}

func (s *StepRun) MarkSkipped() error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusPending {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	s.Status = StatusSkipped
	s.FinishedAt = &now
	s.UpdatedAt = now
	return nil
}

func (s *StepRun) PullEvents() []event.DomainEvent {
	events := s.events
	s.events = nil
	return events
}

func (s *StepRun) recordEvent(e event.DomainEvent) {
	s.events = append(s.events, e)
}

func (s *StepRun) startedEvent(at time.Time) StepRunStarted {
	return StepRunStarted{
		ID:             uuid.New().String(),
		StepRunID:      s.ID.String(),
		WorkflowRunID:  s.WorkflowRunID.String(),
		StepID:         s.StepID.String(),
		WorkflowID:     s.WorkflowID.String(),
		EndpointID:     s.EndpointID.String(),
		OrganizationID: s.OrganizationID.String(),
		Name:           s.Name,
		Description:    s.Description,
		URL:            s.URL,
		Method:         s.Method,
		Headers:        s.Headers,
		Query:          s.Query,
		Body:           s.Body,
		Timeout:        s.Timeout,
		RetryOnFailure: s.RetryOnFailure,
		RetryCount:     s.RetryCount,
		RetryDelay:     s.RetryDelay,
		Index:          s.Index,
		ExecutionOrder: s.ExecutionOrder,
		TreeIndex:      s.TreeIndex,
		Position:       s.Position,
		Status:         string(s.Status),
		Attempt:        s.Attempt,
		Timestamp:      at,
	}
}

func (s *StepRun) succeededEvent(at time.Time) StepRunSucceeded {
	return StepRunSucceeded{
		ID:               uuid.New().String(),
		StepRunID:        s.ID.String(),
		WorkflowRunID:    s.WorkflowRunID.String(),
		StepID:           s.StepID.String(),
		OrganizationID:   s.OrganizationID.String(),
		Status:           string(s.Status),
		Attempt:          s.Attempt,
		ResponseSnapshot: s.ResponseSnapshot,
		Timestamp:        at,
	}
}

func (s *StepRun) failedEvent(at time.Time) StepRunFailed {
	return StepRunFailed{
		ID:               uuid.New().String(),
		StepRunID:        s.ID.String(),
		WorkflowRunID:    s.WorkflowRunID.String(),
		StepID:           s.StepID.String(),
		OrganizationID:   s.OrganizationID.String(),
		Status:           string(s.Status),
		Attempt:          s.Attempt,
		ResponseSnapshot: s.ResponseSnapshot,
		Error:            s.Error,
		Timestamp:        at,
	}
}
