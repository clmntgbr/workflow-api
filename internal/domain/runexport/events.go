package runexport

import (
	"time"

	"go-api/internal/domain/event"
)

const (
	EventTypeRunExportRequested = "runExport.requested.v1"
	EventTypeRunExportReady     = "runExport.ready.v1"
	EventTypeRunExportFailed    = "runExport.failed.v1"
)

type RunExportRequested struct {
	event.PerformedBy
	ID                string    `json:"eventId"`
	RunExportID       string    `json:"runExportId"`
	WorkflowID        string    `json:"workflowId"`
	ProjectID         string    `json:"projectId"`
	RequestedByUserID string    `json:"requestedByUserId"`
	DateRangeFrom     time.Time `json:"dateRangeFrom"`
	DateRangeTo       time.Time `json:"dateRangeTo"`
	Timestamp         time.Time `json:"timestamp"`
}

func (e RunExportRequested) EventID() string       { return e.ID }
func (e RunExportRequested) EventType() string     { return EventTypeRunExportRequested }
func (e RunExportRequested) AggregateID() string   { return e.RunExportID }
func (e RunExportRequested) OccurredAt() time.Time { return e.Timestamp }

type RunExportReady struct {
	ID                string    `json:"eventId"`
	RunExportID       string    `json:"runExportId"`
	WorkflowID        string    `json:"workflowId"`
	ProjectID         string    `json:"projectId"`
	RequestedByUserID string    `json:"requestedByUserId"`
	Timestamp         time.Time `json:"timestamp"`
}

func (e RunExportReady) EventID() string       { return e.ID }
func (e RunExportReady) EventType() string     { return EventTypeRunExportReady }
func (e RunExportReady) AggregateID() string   { return e.RunExportID }
func (e RunExportReady) OccurredAt() time.Time { return e.Timestamp }

type RunExportFailed struct {
	ID                string    `json:"eventId"`
	RunExportID       string    `json:"runExportId"`
	WorkflowID        string    `json:"workflowId"`
	ProjectID         string    `json:"projectId"`
	RequestedByUserID string    `json:"requestedByUserId"`
	Error             string    `json:"error"`
	Timestamp         time.Time `json:"timestamp"`
}

func (e RunExportFailed) EventID() string       { return e.ID }
func (e RunExportFailed) EventType() string     { return EventTypeRunExportFailed }
func (e RunExportFailed) AggregateID() string   { return e.RunExportID }
func (e RunExportFailed) OccurredAt() time.Time { return e.Timestamp }
