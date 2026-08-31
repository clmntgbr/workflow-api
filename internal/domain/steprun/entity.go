package steprun

import (
	"time"

	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/event"
	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type StepRun struct {
	ID            uuid.UUID
	WorkflowRunID uuid.UUID
	StepID        uuid.UUID

	WorkflowID           uuid.UUID
	EndpointID           *uuid.UUID
	ProjectID            uuid.UUID
	StepType             domainstep.Type
	DelayDurationSeconds int
	Name                 string
	Description    string
	URL            string
	Method         string
	Headers        map[string]string
	Query          httpquery.Params
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

	VariableExtracts   []VariableExtract
	Assertions         []domainassertion.Snapshot
	ResponseSnapshot   *ResponseSnapshot
	ExtractedVariables map[string]any
	AssertionsResult   []domainassertion.Result

	StartedAt  *time.Time
	FinishedAt *time.Time
	ResumeAt       *time.Time
	MatchedBranch  *bool
	Error      string
	CreatedAt  time.Time
	UpdatedAt  time.Time

	events []event.DomainEvent
}

type VariableExtract struct {
	VariableID uuid.UUID `json:"variableId"`
	Key        string    `json:"key"`
	Path       string    `json:"path"`
}

type NewStepRunParams struct {
	WorkflowRunID    uuid.UUID
	StepID           uuid.UUID
	WorkflowID           uuid.UUID
	EndpointID           *uuid.UUID
	ProjectID            uuid.UUID
	StepType             domainstep.Type
	DelayDurationSeconds int
	Name                 string
	Description      string
	URL              string
	Method           string
	Headers          map[string]string
	Query            httpquery.Params
	Body             map[string]any
	Timeout          int
	RetryOnFailure   bool
	RetryCount       int
	RetryDelay       int
	Index            string
	ExecutionOrder   int
	TreeIndex        int
	Position         domainstep.Position
	VariableExtracts []VariableExtract
	Assertions       []domainassertion.Snapshot
}

func NewStepRun(p NewStepRunParams) *StepRun {
	now := time.Now().UTC()
	stepType := p.StepType
	if stepType == "" {
		stepType = domainstep.TypeHTTP
	}
	return &StepRun{
		ID:                 uuid.New(),
		WorkflowRunID:      p.WorkflowRunID,
		StepID:             p.StepID,
		WorkflowID:           p.WorkflowID,
		EndpointID:           p.EndpointID,
		ProjectID:            p.ProjectID,
		StepType:             stepType,
		DelayDurationSeconds: p.DelayDurationSeconds,
		Name:                 p.Name,
		Description:        p.Description,
		URL:                p.URL,
		Method:             p.Method,
		Headers:            normalizeStringMap(p.Headers),
		Query:              httpquery.Clone(p.Query),
		Body:               normalizeAnyMap(p.Body),
		Timeout:            p.Timeout,
		RetryOnFailure:     p.RetryOnFailure,
		RetryCount:         p.RetryCount,
		RetryDelay:         p.RetryDelay,
		Index:              p.Index,
		ExecutionOrder:     p.ExecutionOrder,
		TreeIndex:          p.TreeIndex,
		Position:           p.Position,
		VariableExtracts:   append([]VariableExtract(nil), p.VariableExtracts...),
		Assertions:         append([]domainassertion.Snapshot(nil), p.Assertions...),
		ExtractedVariables: map[string]any{},
		AssertionsResult:   []domainassertion.Result{},
		Status:             StatusPending,
		Attempt:            0,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

func (s *StepRun) Queue() {
	s.recordEvent(StepRunQueued{
		ID:            uuid.New().String(),
		StepRunID:     s.ID.String(),
		WorkflowRunID: s.WorkflowRunID.String(),
		StepID:        s.StepID.String(),
		ProjectID:     s.ProjectID.String(),
		Timestamp:     time.Now().UTC(),
	})
}

func (s *StepRun) MarkStarted() error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusPending && s.Status != StatusWaiting {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	s.Status = StatusRunning
	if s.Attempt == 0 {
		s.Attempt = 1
	}
	if s.StartedAt == nil {
		s.StartedAt = &now
	}
	s.UpdatedAt = now
	s.recordEvent(s.startedEvent(now))
	return nil
}

func (s *StepRun) IncrementAttempt() error {
	if s.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}
	s.Attempt++
	s.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *StepRun) CanRetry() bool {
	return s.RetryOnFailure && s.Attempt < s.RetryCount
}

func (s *StepRun) MarkSucceeded(
	response ResponseSnapshot,
	extracted map[string]any,
	assertionsResult []domainassertion.Result,
) error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusRunning {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	normalized := response.Normalized()
	s.Status = StatusSuccess
	s.ResponseSnapshot = &normalized
	if extracted == nil {
		extracted = map[string]any{}
	}
	s.ExtractedVariables = extracted
	if assertionsResult == nil {
		assertionsResult = []domainassertion.Result{}
	}
	s.AssertionsResult = assertionsResult
	s.Error = ""
	s.FinishedAt = &now
	s.UpdatedAt = now
	s.recordEvent(s.succeededEvent(now))
	return nil
}

func (s *StepRun) MarkConditionSucceeded(matchedBranch bool) error {
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
	s.Status = StatusSuccess
	s.MatchedBranch = &matchedBranch
	s.Error = ""
	s.FinishedAt = &now
	s.UpdatedAt = now
	s.recordEvent(s.succeededEvent(now))
	return nil
}

func (s *StepRun) MarkFailed(
	errMsg string,
	response *ResponseSnapshot,
	assertionsResult []domainassertion.Result,
) error {
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
		normalized := response.Normalized()
		s.ResponseSnapshot = &normalized
	}
	if assertionsResult == nil {
		s.AssertionsResult = []domainassertion.Result{}
	} else {
		s.AssertionsResult = assertionsResult
	}
	s.Status = StatusFailed
	s.Error = errMsg
	s.FinishedAt = &now
	s.UpdatedAt = now
	s.recordEvent(s.failedEvent(now))
	return nil
}

func (s *StepRun) MarkWaiting(resumeAt time.Time) error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusPending {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	s.Status = StatusWaiting
	s.ResumeAt = &resumeAt
	if s.StartedAt == nil {
		s.StartedAt = &now
	}
	s.UpdatedAt = now
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

func (s *StepRun) MarkCancelled() error {
	if s.Status.IsTerminal() {
		return ErrAlreadyTerminal
	}
	if s.Status != StatusPending && s.Status != StatusRunning && s.Status != StatusWaiting {
		return ErrInvalidStatusTransition
	}

	now := time.Now().UTC()
	s.Status = StatusCancelled
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
		EndpointID:     endpointIDString(s.EndpointID),
		ProjectID:      s.ProjectID.String(),
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
		ID:                 uuid.New().String(),
		StepRunID:          s.ID.String(),
		WorkflowRunID:      s.WorkflowRunID.String(),
		StepID:             s.StepID.String(),
		ProjectID:          s.ProjectID.String(),
		Status:             string(s.Status),
		Attempt:            s.Attempt,
		ResponseSnapshot:   s.ResponseSnapshot,
		ExtractedVariables: s.ExtractedVariables,
		Timestamp:          at,
	}
}

func (s *StepRun) failedEvent(at time.Time) StepRunFailed {
	return StepRunFailed{
		ID:               uuid.New().String(),
		StepRunID:        s.ID.String(),
		WorkflowRunID:    s.WorkflowRunID.String(),
		StepID:           s.StepID.String(),
		ProjectID:        s.ProjectID.String(),
		Status:           string(s.Status),
		Attempt:          s.Attempt,
		ResponseSnapshot: s.ResponseSnapshot,
		Error:            s.Error,
		Timestamp:        at,
	}
}

func endpointIDString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}
