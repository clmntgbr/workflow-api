package project

import "time"

const (
	EventTypeProjectCreated       = "project.created.v1"
	EventTypeProjectUpdated       = "project.updated.v1"
	EventTypeProjectDeleted       = "project.deleted.v1"
	EventTypeProjectMemberAdded   = "project.member_added.v1"
	EventTypeProjectMemberRemoved = "project.member_removed.v1"
)

type ProjectCreated struct {
	ID              string    `json:"eventId"`
	ProjectID  string    `json:"projectId"`
	Name            string    `json:"name"`
	CreatedByUserID string    `json:"createdByUserId"`
	Timestamp       time.Time `json:"timestamp"`
}

func (e ProjectCreated) EventID() string       { return e.ID }
func (e ProjectCreated) EventType() string     { return EventTypeProjectCreated }
func (e ProjectCreated) AggregateID() string   { return e.ProjectID }
func (e ProjectCreated) OccurredAt() time.Time { return e.Timestamp }

type ProjectUpdated struct {
	ID             string    `json:"eventId"`
	ProjectID string    `json:"projectId"`
	Name           string    `json:"name"`
	MemberIDs      []string  `json:"memberIds"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e ProjectUpdated) EventID() string       { return e.ID }
func (e ProjectUpdated) EventType() string     { return EventTypeProjectUpdated }
func (e ProjectUpdated) AggregateID() string   { return e.ProjectID }
func (e ProjectUpdated) OccurredAt() time.Time { return e.Timestamp }

type ProjectDeleted struct {
	ID             string    `json:"eventId"`
	ProjectID string    `json:"projectId"`
	MemberIDs      []string  `json:"memberIds"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e ProjectDeleted) EventID() string       { return e.ID }
func (e ProjectDeleted) EventType() string     { return EventTypeProjectDeleted }
func (e ProjectDeleted) AggregateID() string   { return e.ProjectID }
func (e ProjectDeleted) OccurredAt() time.Time { return e.Timestamp }

type ProjectMemberAdded struct {
	ID             string    `json:"eventId"`
	ProjectID string    `json:"projectId"`
	UserID         string    `json:"userId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e ProjectMemberAdded) EventID() string       { return e.ID }
func (e ProjectMemberAdded) EventType() string     { return EventTypeProjectMemberAdded }
func (e ProjectMemberAdded) AggregateID() string   { return e.ProjectID }
func (e ProjectMemberAdded) OccurredAt() time.Time { return e.Timestamp }

type ProjectMemberRemoved struct {
	ID             string    `json:"eventId"`
	ProjectID string    `json:"projectId"`
	UserID         string    `json:"userId"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e ProjectMemberRemoved) EventID() string       { return e.ID }
func (e ProjectMemberRemoved) EventType() string     { return EventTypeProjectMemberRemoved }
func (e ProjectMemberRemoved) AggregateID() string   { return e.ProjectID }
func (e ProjectMemberRemoved) OccurredAt() time.Time { return e.Timestamp }
