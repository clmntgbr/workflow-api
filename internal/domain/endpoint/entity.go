package endpoint

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Endpoint struct {
	ID             uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         Method
	Headers        map[string]string
	Query          map[string]string
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Status         Status
	OrganizationID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time

	events []event.DomainEvent
}

type NewEndpointParams struct {
	Name             string
	Description      string
	URL              string
	Method           Method
	Headers          map[string]string
	Query            map[string]string
	Body             map[string]any
	Timeout          int
	RetryOnFailure   bool
	RetryCount       int
	RetryDelay       int
	Status           Status
	SkipCreatedEvent bool
	OrganizationID   uuid.UUID
}

func NewEndpoint(p NewEndpointParams) *Endpoint {
	now := time.Now().UTC()
	headers := p.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	query := p.Query
	if query == nil {
		query = map[string]string{}
	}
	body := p.Body
	if body == nil {
		body = map[string]any{}
	}
	status := p.Status
	if status == "" {
		status = StatusActive
	}

	e := &Endpoint{
		ID:             uuid.New(),
		Name:           p.Name,
		Description:    p.Description,
		URL:            p.URL,
		Method:         p.Method,
		Headers:        headers,
		Query:          query,
		Body:           body,
		Timeout:        p.Timeout,
		RetryOnFailure: p.RetryOnFailure,
		RetryCount:     p.RetryCount,
		RetryDelay:     p.RetryDelay,
		Status:         status,
		OrganizationID: p.OrganizationID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if !p.SkipCreatedEvent {
		e.recordEvent(EndpointCreated{
			ID:             uuid.New().String(),
			EndpointID:     e.ID.String(),
			OrganizationID: e.OrganizationID.String(),
			Name:           e.Name,
			Description:    e.Description,
			URL:            e.URL,
			Method:         string(e.Method),
			Headers:        e.Headers,
			Query:          e.Query,
			Body:           e.Body,
			Timeout:        e.Timeout,
			RetryOnFailure: e.RetryOnFailure,
			RetryCount:     e.RetryCount,
			RetryDelay:     e.RetryDelay,
			Status:         string(e.Status),
			Timestamp:      now,
		})
	}
	return e
}

func (e *Endpoint) PullEvents() []event.DomainEvent {
	events := e.events
	e.events = nil
	return events
}

type UpdateEndpointParams struct {
	Name           string
	Description    string
	URL            string
	Method         Method
	Headers        map[string]string
	Query          map[string]string
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
	Status         Status
}

func (e *Endpoint) ApplyUpdate(p UpdateEndpointParams) {
	headers := p.Headers
	if headers == nil {
		headers = map[string]string{}
	}
	query := p.Query
	if query == nil {
		query = map[string]string{}
	}
	body := p.Body
	if body == nil {
		body = map[string]any{}
	}

	e.Name = p.Name
	e.Description = p.Description
	e.URL = p.URL
	e.Method = p.Method
	e.Headers = headers
	e.Query = query
	e.Body = body
	e.Timeout = p.Timeout
	e.RetryOnFailure = p.RetryOnFailure
	e.RetryCount = p.RetryCount
	e.RetryDelay = p.RetryDelay
	e.Status = p.Status
	e.UpdatedAt = time.Now().UTC()

	e.recordEvent(EndpointUpdated{
		ID:             uuid.New().String(),
		EndpointID:     e.ID.String(),
		OrganizationID: e.OrganizationID.String(),
		Name:           e.Name,
		Description:    e.Description,
		URL:            e.URL,
		Method:         string(e.Method),
		Headers:        e.Headers,
		Query:          e.Query,
		Body:           e.Body,
		Timeout:        e.Timeout,
		RetryOnFailure: e.RetryOnFailure,
		RetryCount:     e.RetryCount,
		RetryDelay:     e.RetryDelay,
		Status:         string(e.Status),
		Timestamp:      e.UpdatedAt,
	})
}

func (e *Endpoint) MarkDeleted() {
	e.Status = StatusDeleted
	e.UpdatedAt = time.Now().UTC()
	e.recordEvent(EndpointDeleted{
		ID:             uuid.New().String(),
		EndpointID:     e.ID.String(),
		OrganizationID: e.OrganizationID.String(),
		Timestamp:      e.UpdatedAt,
	})
}

func (e *Endpoint) recordEvent(evt event.DomainEvent) {
	e.events = append(e.events, evt)
}
