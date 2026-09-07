package runexport

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type RunExport struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	ProjectID         uuid.UUID
	RequestedByUserID uuid.UUID
	DateRangeFrom     time.Time
	DateRangeTo       time.Time
	Status            Status
	Error             string
	CreatedAt         time.Time
	UpdatedAt         time.Time

	events []event.DomainEvent
}

type NewRunExportParams struct {
	WorkflowID        uuid.UUID
	ProjectID         uuid.UUID
	RequestedByUserID uuid.UUID
	DateRangeFrom     time.Time
	DateRangeTo       time.Time
}

func NewRunExport(p NewRunExportParams) *RunExport {
	now := time.Now().UTC()
	job := &RunExport{
		ID:                uuid.New(),
		WorkflowID:        p.WorkflowID,
		ProjectID:         p.ProjectID,
		RequestedByUserID: p.RequestedByUserID,
		DateRangeFrom:     p.DateRangeFrom.UTC(),
		DateRangeTo:       p.DateRangeTo.UTC(),
		Status:            StatusPending,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	job.recordEvent(RunExportRequested{
		ID:                uuid.New().String(),
		RunExportID:       job.ID.String(),
		WorkflowID:        job.WorkflowID.String(),
		ProjectID:         job.ProjectID.String(),
		RequestedByUserID: job.RequestedByUserID.String(),
		DateRangeFrom:     job.DateRangeFrom,
		DateRangeTo:       job.DateRangeTo,
		Timestamp:         now,
	})
	return job
}

func (j *RunExport) PullEvents() []event.DomainEvent {
	events := j.events
	j.events = nil
	return events
}

func (j *RunExport) recordEvent(e event.DomainEvent) {
	j.events = append(j.events, e)
}

func (j *RunExport) MarkProcessing() {
	if j.Status == StatusReady || j.Status == StatusFailed {
		return
	}
	j.Status = StatusProcessing
	j.Error = ""
	j.UpdatedAt = time.Now().UTC()
}

func (j *RunExport) MarkReady() {
	if j.Status == StatusReady {
		return
	}
	now := time.Now().UTC()
	j.Status = StatusReady
	j.Error = ""
	j.UpdatedAt = now
	j.recordEvent(RunExportReady{
		ID:                uuid.New().String(),
		RunExportID:       j.ID.String(),
		WorkflowID:        j.WorkflowID.String(),
		ProjectID:         j.ProjectID.String(),
		RequestedByUserID: j.RequestedByUserID.String(),
		Timestamp:         now,
	})
}

func (j *RunExport) MarkFailed(reason string) {
	if j.Status == StatusReady {
		return
	}
	now := time.Now().UTC()
	j.Status = StatusFailed
	j.Error = reason
	j.UpdatedAt = now
	j.recordEvent(RunExportFailed{
		ID:                uuid.New().String(),
		RunExportID:       j.ID.String(),
		WorkflowID:        j.WorkflowID.String(),
		ProjectID:         j.ProjectID.String(),
		RequestedByUserID: j.RequestedByUserID.String(),
		Error:             reason,
		Timestamp:         now,
	})
}
