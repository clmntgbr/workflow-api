package endpoint

import "time"

const (
	EventTypeEndpointCreated = "endpoint.created.v1"
	EventTypeEndpointUpdated = "endpoint.updated.v1"
)

type EndpointCreated struct {
	ID             string            `json:"eventId"`
	EndpointID     string            `json:"endpointId"`
	OrganizationID string            `json:"organizationId"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Query          map[string]string `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
	Status         string            `json:"status"`
	Timestamp      time.Time         `json:"timestamp"`
}

func (e EndpointCreated) EventID() string       { return e.ID }
func (e EndpointCreated) EventType() string     { return EventTypeEndpointCreated }
func (e EndpointCreated) AggregateID() string   { return e.EndpointID }
func (e EndpointCreated) OccurredAt() time.Time { return e.Timestamp }

type EndpointUpdated struct {
	ID             string            `json:"eventId"`
	EndpointID     string            `json:"endpointId"`
	OrganizationID string            `json:"organizationId"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Query          map[string]string `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
	Status         string            `json:"status"`
	Timestamp      time.Time         `json:"timestamp"`
}

func (e EndpointUpdated) EventID() string       { return e.ID }
func (e EndpointUpdated) EventType() string     { return EventTypeEndpointUpdated }
func (e EndpointUpdated) AggregateID() string   { return e.EndpointID }
func (e EndpointUpdated) OccurredAt() time.Time { return e.Timestamp }

const EventTypeEndpointDeleted = "endpoint.deleted.v1"

type EndpointDeleted struct {
	ID             string    `json:"eventId"`
	EndpointID     string    `json:"endpointId"`
	OrganizationID string    `json:"organizationId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e EndpointDeleted) EventID() string       { return e.ID }
func (e EndpointDeleted) EventType() string     { return EventTypeEndpointDeleted }
func (e EndpointDeleted) AggregateID() string   { return e.EndpointID }
func (e EndpointDeleted) OccurredAt() time.Time { return e.Timestamp }
