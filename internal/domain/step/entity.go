package step

import (
	"strings"
	"time"

	"go-api/internal/domain/event"
	"go-api/internal/domain/httpquery"

	"github.com/google/uuid"
)

type Step struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	EndpointID     *uuid.UUID
	ProjectID      uuid.UUID
	Type                 Type
	DelayDurationSeconds int
	Expression           *string

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
	ProjectID      uuid.UUID
	Endpoint       EndpointSnapshot
	Index          string
	ExecutionOrder int
	TreeIndex      int
	Position       Position
}

type NewDelayStepParams struct {
	ID                   uuid.UUID
	WorkflowID           uuid.UUID
	ProjectID            uuid.UUID
	Name                 string
	DelayDurationSeconds int
	Index                string
	ExecutionOrder       int
	TreeIndex            int
	Position             Position
}

type NewConditionStepParams struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	Expression     string
	Index          string
	ExecutionOrder int
	TreeIndex      int
	Position       Position
}

func NewDelayStep(p NewDelayStepParams) (*Step, error) {
	now := time.Now().UTC()
	id := p.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	name := p.Name
	if name == "" {
		name = "Delay"
	}

	s := &Step{
		ID:                   id,
		WorkflowID:           p.WorkflowID,
		EndpointID:           nil,
		ProjectID:            p.ProjectID,
		Type:                 TypeDelay,
		DelayDurationSeconds: p.DelayDurationSeconds,
		Name:                 name,
		Headers:              map[string]string{},
		Query:                httpquery.Empty(),
		Body:                 map[string]any{},
		Index:                p.Index,
		ExecutionOrder:       p.ExecutionOrder,
		TreeIndex:            p.TreeIndex,
		Position:             p.Position,
		Status:               StatusActive,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := ValidateConfig(s); err != nil {
		return nil, err
	}
	s.recordEvent(s.newCreatedEvent(now))
	return s, nil
}

func NewConditionStep(p NewConditionStepParams) (*Step, error) {
	now := time.Now().UTC()
	id := p.ID
	if id == uuid.Nil {
		id = uuid.New()
	}
	name := p.Name
	if name == "" {
		name = "Condition"
	}
	expression := strings.TrimSpace(p.Expression)

	s := &Step{
		ID:             id,
		WorkflowID:     p.WorkflowID,
		EndpointID:     nil,
		ProjectID:      p.ProjectID,
		Type:           TypeCondition,
		Expression:     &expression,
		Name:           name,
		Headers:        map[string]string{},
		Query:          httpquery.Empty(),
		Body:           map[string]any{},
		Index:          p.Index,
		ExecutionOrder: p.ExecutionOrder,
		TreeIndex:      p.TreeIndex,
		Position:       p.Position,
		Status:         StatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := ValidateConfig(s); err != nil {
		return nil, err
	}
	s.recordEvent(s.newCreatedEvent(now))
	return s, nil
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

	endpointID := p.EndpointID
	s := &Step{
		ID:             id,
		WorkflowID:     p.WorkflowID,
		EndpointID:     &endpointID,
		ProjectID:      p.ProjectID,
		Type:           TypeHTTP,
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
	s.recordEvent(s.newCreatedEvent(now))
	return s
}

func (s *Step) newCreatedEvent(at time.Time) StepCreated {
	return StepCreated{
		ID:                   uuid.New().String(),
		StepID:               s.ID.String(),
		WorkflowID:           s.WorkflowID.String(),
		EndpointID:           endpointIDString(s.EndpointID),
		ProjectID:            s.ProjectID.String(),
		Type:                 string(s.Type),
		DelayDurationSeconds: s.DelayDurationSeconds,
		Expression:           s.Expression,
		Name:                 s.Name,
		Description:          s.Description,
		URL:                  s.URL,
		Method:               s.Method,
		Headers:              s.Headers,
		Query:                s.Query,
		Body:                 s.Body,
		Timeout:              s.Timeout,
		RetryOnFailure:       s.RetryOnFailure,
		RetryCount:           s.RetryCount,
		RetryDelay:           s.RetryDelay,
		Index:                s.Index,
		ExecutionOrder:       s.ExecutionOrder,
		TreeIndex:            s.TreeIndex,
		Position:             s.Position,
		Status:               string(s.Status),
		Timestamp:            at,
	}
}

func endpointIDString(id *uuid.UUID) string {
	if id == nil {
		return ""
	}
	return id.String()
}

func (s *Step) ApplyPositionUpdate(index string, position Position) {
	s.Index = index
	s.ExecutionOrder = CalculateExecutionOrder(index)
	s.Position = position
	s.UpdatedAt = time.Now().UTC()
	s.recordPositionUpdatedEvent()
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

type UpdateDelayStepConfigParams struct {
	Name                 string
	Description          string
	DelayDurationSeconds int
}

type UpdateConditionStepConfigParams struct {
	Name        string
	Description string
	Expression  string
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

func (s *Step) ApplyDelayConfigUpdate(p UpdateDelayStepConfigParams) error {
	if s.Type != TypeDelay {
		return ErrInvalidStepTypeConfig
	}
	s.Name = p.Name
	s.Description = p.Description
	s.DelayDurationSeconds = p.DelayDurationSeconds
	s.UpdatedAt = time.Now().UTC()
	if err := ValidateConfig(s); err != nil {
		return err
	}
	s.recordUpdatedEvent()
	return nil
}

func (s *Step) ApplyConditionConfigUpdate(p UpdateConditionStepConfigParams) error {
	if s.Type != TypeCondition {
		return ErrInvalidStepTypeConfig
	}
	expression := strings.TrimSpace(p.Expression)
	s.Name = p.Name
	s.Description = p.Description
	s.Expression = &expression
	s.UpdatedAt = time.Now().UTC()
	if err := ValidateConfig(s); err != nil {
		return err
	}
	s.recordUpdatedEvent()
	return nil
}

func (s *Step) recordUpdatedEvent() {
	s.recordEvent(StepUpdated{
		ID:                   uuid.New().String(),
		StepID:               s.ID.String(),
		WorkflowID:           s.WorkflowID.String(),
		ProjectID:            s.ProjectID.String(),
		Type:                 string(s.Type),
		DelayDurationSeconds: s.DelayDurationSeconds,
		Expression:           s.Expression,
		Name:                 s.Name,
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

func (s *Step) recordPositionUpdatedEvent() {
	s.recordEvent(StepPositionUpdated{
		ID:             uuid.New().String(),
		StepID:         s.ID.String(),
		WorkflowID:     s.WorkflowID.String(),
		ProjectID:      s.ProjectID.String(),
		Name:           s.Name,
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
