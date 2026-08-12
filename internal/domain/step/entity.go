package step

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Step struct {
	ID             uuid.UUID
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
	Query          map[string]string
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
	OrganizationID uuid.UUID
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
	query := p.Endpoint.Query
	if query == nil {
		query = map[string]string{}
	}
	body := p.Endpoint.Body
	if body == nil {
		body = map[string]any{}
	}

	s := &Step{
		ID:             id,
		WorkflowID:     p.WorkflowID,
		EndpointID:     p.EndpointID,
		OrganizationID: p.OrganizationID,
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
		Timestamp:      now,
	})
	return s
}

func (s *Step) ApplyPositionUpdate(index string, position Position) {
	s.Index = index
	s.ExecutionOrder = CalculateExecutionOrder(index)
	s.Position = position
	s.UpdatedAt = time.Now().UTC()

	s.recordEvent(StepUpdated{
		ID:             uuid.New().String(),
		StepID:         s.ID.String(),
		WorkflowID:     s.WorkflowID.String(),
		OrganizationID: s.OrganizationID.String(),
		Index:          s.Index,
		ExecutionOrder: s.ExecutionOrder,
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
		OrganizationID: s.OrganizationID.String(),
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
