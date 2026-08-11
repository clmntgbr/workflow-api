package organization

import "time"

const (
	EventTypeOrganizationCreated       = "organization.created.v1"
	EventTypeOrganizationUpdated       = "organization.updated.v1"
	EventTypeOrganizationDeleted       = "organization.deleted.v1"
	EventTypeOrganizationMemberAdded   = "organization.member_added.v1"
	EventTypeOrganizationMemberRemoved = "organization.member_removed.v1"
)

type OrganizationCreated struct {
	ID             string    `json:"eventId"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e OrganizationCreated) EventID() string       { return e.ID }
func (e OrganizationCreated) EventType() string     { return EventTypeOrganizationCreated }
func (e OrganizationCreated) AggregateID() string   { return e.OrganizationID }
func (e OrganizationCreated) OccurredAt() time.Time { return e.Timestamp }

type OrganizationUpdated struct {
	ID             string    `json:"eventId"`
	OrganizationID string    `json:"organizationId"`
	Name           string    `json:"name"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e OrganizationUpdated) EventID() string       { return e.ID }
func (e OrganizationUpdated) EventType() string     { return EventTypeOrganizationUpdated }
func (e OrganizationUpdated) AggregateID() string   { return e.OrganizationID }
func (e OrganizationUpdated) OccurredAt() time.Time { return e.Timestamp }

type OrganizationDeleted struct {
	ID             string    `json:"eventId"`
	OrganizationID string    `json:"organizationId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e OrganizationDeleted) EventID() string       { return e.ID }
func (e OrganizationDeleted) EventType() string     { return EventTypeOrganizationDeleted }
func (e OrganizationDeleted) AggregateID() string   { return e.OrganizationID }
func (e OrganizationDeleted) OccurredAt() time.Time { return e.Timestamp }

type OrganizationMemberAdded struct {
	ID             string    `json:"eventId"`
	OrganizationID string    `json:"organizationId"`
	UserID         string    `json:"userId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e OrganizationMemberAdded) EventID() string       { return e.ID }
func (e OrganizationMemberAdded) EventType() string     { return EventTypeOrganizationMemberAdded }
func (e OrganizationMemberAdded) AggregateID() string   { return e.OrganizationID }
func (e OrganizationMemberAdded) OccurredAt() time.Time { return e.Timestamp }

type OrganizationMemberRemoved struct {
	ID             string    `json:"eventId"`
	OrganizationID string    `json:"organizationId"`
	UserID         string    `json:"userId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e OrganizationMemberRemoved) EventID() string       { return e.ID }
func (e OrganizationMemberRemoved) EventType() string     { return EventTypeOrganizationMemberRemoved }
func (e OrganizationMemberRemoved) AggregateID() string   { return e.OrganizationID }
func (e OrganizationMemberRemoved) OccurredAt() time.Time { return e.Timestamp }
