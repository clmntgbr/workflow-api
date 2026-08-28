package activitylog

import (
	"time"

	"github.com/google/uuid"
)

type Entry struct {
	ID              uuid.UUID
	ProjectID       uuid.UUID
	Action          string
	SubjectType     string
	SubjectID       uuid.UUID
	WorkflowID      *uuid.UUID
	WorkflowRunID   *uuid.UUID
	StepID          *uuid.UUID
	StepRunID       *uuid.UUID
	ActorType       string
	ActorUserID     *uuid.UUID
	Level           string
	Message         string
	Payload         map[string]any
	SourceEventID   uuid.UUID
	SourceEventType string
	OccurredAt      time.Time
	CreatedAt       time.Time
}

type View struct {
	ID              uuid.UUID
	ProjectID       uuid.UUID
	Action          string
	SubjectType     string
	SubjectID       uuid.UUID
	WorkflowID      *uuid.UUID
	WorkflowRunID   *uuid.UUID
	StepID          *uuid.UUID
	StepRunID       *uuid.UUID
	ActorType       string
	ActorUserID     *uuid.UUID
	Level           string
	Message         string
	Payload         map[string]any
	SourceEventID   uuid.UUID
	SourceEventType string
	OccurredAt      time.Time
	CreatedAt       time.Time
}
