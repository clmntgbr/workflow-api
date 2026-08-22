package user

import "time"

const (
	EventTypeUserCreated                   = "user.created.v1"
	EventTypeUserUpdated                   = "user.updated.v1"
	EventTypeUserDeleted                   = "user.deleted.v1"
	EventTypeUserActiveProjectChanged = "user.active_project_changed.v1"
)

type UserCreated struct {
	ID        string    `json:"eventId"`
	UserID    string    `json:"userId"`
	ClerkID   string    `json:"clerkId"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Banned    bool      `json:"banned"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserCreated) EventID() string       { return e.ID }
func (e UserCreated) EventType() string     { return EventTypeUserCreated }
func (e UserCreated) AggregateID() string   { return e.UserID }
func (e UserCreated) OccurredAt() time.Time { return e.Timestamp }

type UserUpdated struct {
	ID        string    `json:"eventId"`
	UserID    string    `json:"userId"`
	ClerkID   string    `json:"clerkId"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Banned    bool      `json:"banned"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserUpdated) EventID() string       { return e.ID }
func (e UserUpdated) EventType() string     { return EventTypeUserUpdated }
func (e UserUpdated) AggregateID() string   { return e.UserID }
func (e UserUpdated) OccurredAt() time.Time { return e.Timestamp }

type UserDeleted struct {
	ID        string    `json:"eventId"`
	UserID    string    `json:"userId"`
	ClerkID   string    `json:"clerkId"`
	Timestamp time.Time `json:"timestamp"`
}

func (e UserDeleted) EventID() string       { return e.ID }
func (e UserDeleted) EventType() string     { return EventTypeUserDeleted }
func (e UserDeleted) AggregateID() string   { return e.UserID }
func (e UserDeleted) OccurredAt() time.Time { return e.Timestamp }

type UserActiveProjectChanged struct {
	ID             string    `json:"eventId"`
	UserID         string    `json:"userId"`
	ProjectID string    `json:"projectId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e UserActiveProjectChanged) EventID() string { return e.ID }
func (e UserActiveProjectChanged) EventType() string {
	return EventTypeUserActiveProjectChanged
}
func (e UserActiveProjectChanged) AggregateID() string   { return e.UserID }
func (e UserActiveProjectChanged) OccurredAt() time.Time { return e.Timestamp }
