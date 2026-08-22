package step

import (
	"time"

	"go-api/internal/domain/event"
	"go-api/internal/domain/httpquery"

	"github.com/google/uuid"
)

type Step struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	EndpointID     uuid.UUID
	ProjectID uuid.UUID

	Name           string
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
	Position       Position

	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

type EndpointSnapshot struct {
	ID             uuid.UUID
	Name           string
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
}

type NewStepParams struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	EndpointID     uuid.UUID
	ProjectID uuid.UUID
	Endpoint       EndpointSnapshot
	Index          string
	ExecutionOrder int
	TreeIndex      int
	Position       Position
}

func NewStep(p NewStepParams) *Step {
	now := time.Now().UTC()
	id := p.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	headers := p.Endpoint.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	query := httpquery.Clone(p.Endpoint.Query)
	body := p.Endpoint.Body
	if body == nil {
		body = map[string]any{}
	}

	s := &Step{
		ID:             id,
		WorkflowID:     p.WorkflowID,
		EndpointID:     p.EndpointID,
		ProjectID: p.ProjectID,
		Name:           p.Endpoint.Name,
		Description:    p.Endpoint.Description,
		URL:            p.Endpoint.URL,
		Method:         p.Endpoint.Method,
		Headers:        headers,
		Query:          query,
		Body:           body,
		Timeout:        p.Endpoint.Timeout,
		RetryOnFailure: p.Endpoint.RetryOnFailure,
		RetryCount:     p.Endpoint.RetryCount,
		RetryDelay:     p.Endpoint.RetryDelay,
		Index:          p.Index,
		ExecutionOrder: p.ExecutionOrder,
		TreeIndex:      p.TreeIndex,
		Position:       p.Position,
		Status:         StatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	s.recordEvent(StepCreated{
		ID:             uuid.New().String(),
		StepID:         s.ID.String(),
		WorkflowID:     s.WorkflowID.String(),
		EndpointID:     s.EndpointID.String(),
		ProjectID: s.ProjectID.String(),
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
		Timestamp:      now,
	})
	return s
}

func (s *Step) ApplyPositionUpdate(index string, position Position) {
	s.Index = index
	s.ExecutionOrder = CalculateExecutionOrder(index)
	s.Position = position
	s.UpdatedAt = time.Now().UTC()
	s.recordUpdatedEvent()
}

type UpdateStepConfigParams struct {
	Name           string
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
}

func (s *Step) ApplyConfigUpdate(p UpdateStepConfigParams) {
	headers := p.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	query := httpquery.Clone(p.Query)
	body := p.Body
	if body == nil {
		body = map[string]any{}
	}

	s.Name = p.Name
	s.Description = p.Description
	s.URL = p.URL
	s.Method = p.Method
	s.Headers = headers
	s.Query = query
	s.Body = body
	s.Timeout = p.Timeout
	s.RetryOnFailure = p.RetryOnFailure
	s.RetryCount = p.RetryCount
	s.RetryDelay = p.RetryDelay
	s.UpdatedAt = time.Now().UTC()
	s.recordUpdatedEvent()
}

func (s *Step) recordUpdatedEvent() {
	s.recordEvent(StepUpdated{
		ID:             uuid.New().String(),
		StepID:         s.ID.String(),
		WorkflowID:     s.WorkflowID.String(),
		ProjectID: s.ProjectID.String(),
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
		Timestamp:      s.UpdatedAt,
	})
}

func (s *Step) MarkDeleted() {
	s.Status = StatusDeleted
	s.UpdatedAt = time.Now().UTC()
	s.recordEvent(StepDeleted{
		ID:             uuid.New().String(),
		StepID:         s.ID.String(),
		WorkflowID:     s.WorkflowID.String(),
		ProjectID: s.ProjectID.String(),
		Timestamp:      s.UpdatedAt,
	})
}

func (s *Step) PullEvents() []event.DomainEvent {
	events := s.events
	s.events = nil
	return events
}

func (s *Step) recordEvent(evt event.DomainEvent) {
	s.events = append(s.events, evt)
}
