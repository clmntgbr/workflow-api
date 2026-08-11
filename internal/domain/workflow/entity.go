package workflow

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Workflow struct {
	ID          uuid.UUID
	Name        string
	Description string
	Status      Status

	OrganizationID uuid.UUID

	ScheduleIntervalMinutes int
	Concurrency             int

	NotificationsEnabled bool
	NotifyOnSuccess      bool
	NotifyOnFailure      bool
	NotifyOnCancel       bool

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

type NewWorkflowParams struct {
	Name                    string
	Description             string
	OrganizationID          uuid.UUID
	ScheduleIntervalMinutes int
	Concurrency             int
	NotificationsEnabled    bool
	NotifyOnSuccess         bool
	NotifyOnFailure         bool
	NotifyOnCancel          bool
}

func NewWorkflow(p NewWorkflowParams) *Workflow {
	now := time.Now().UTC()
	concurrency := p.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	w := &Workflow{
		ID:                      uuid.New(),
		Name:                    p.Name,
		Description:             p.Description,
		Status:                  StatusInactive,
		OrganizationID:          p.OrganizationID,
		ScheduleIntervalMinutes: p.ScheduleIntervalMinutes,
		Concurrency:             concurrency,
		NotificationsEnabled:    p.NotificationsEnabled,
		NotifyOnSuccess:         p.NotifyOnSuccess,
		NotifyOnFailure:         p.NotifyOnFailure,
		NotifyOnCancel:          p.NotifyOnCancel,
		CreatedAt:               now,
		UpdatedAt:               now,
	}
	w.recordEvent(WorkflowCreated{
		ID:                      uuid.New().String(),
		WorkflowID:              w.ID.String(),
		OrganizationID:          w.OrganizationID.String(),
		Name:                    w.Name,
		Description:             w.Description,
		Status:                  string(w.Status),
		ScheduleIntervalMinutes: w.ScheduleIntervalMinutes,
		Concurrency:             w.Concurrency,
		NotificationsEnabled:    w.NotificationsEnabled,
		NotifyOnSuccess:         w.NotifyOnSuccess,
		NotifyOnFailure:         w.NotifyOnFailure,
		NotifyOnCancel:          w.NotifyOnCancel,
		Timestamp:               now,
	})
	return w
}

func (w *Workflow) PullEvents() []event.DomainEvent {
	events := w.events
	w.events = nil
	return events
}

func (w *Workflow) recordEvent(e event.DomainEvent) {
	w.events = append(w.events, e)
}

type UpdateWorkflowParams struct {
	Name                    string
	Description             string
	Status                  Status
	ScheduleIntervalMinutes int
	Concurrency             int
	NotificationsEnabled    bool
	NotifyOnSuccess         bool
	NotifyOnFailure         bool
	NotifyOnCancel          bool
}

func (w *Workflow) ApplyUpdate(p UpdateWorkflowParams) {
	concurrency := p.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	w.Name = p.Name
	w.Description = p.Description
	w.Status = p.Status
	w.ScheduleIntervalMinutes = p.ScheduleIntervalMinutes
	w.Concurrency = concurrency
	w.NotificationsEnabled = p.NotificationsEnabled
	w.NotifyOnSuccess = p.NotifyOnSuccess
	w.NotifyOnFailure = p.NotifyOnFailure
	w.NotifyOnCancel = p.NotifyOnCancel
	w.UpdatedAt = time.Now().UTC()

	w.recordEvent(WorkflowUpdated{
		ID:                      uuid.New().String(),
		WorkflowID:              w.ID.String(),
		OrganizationID:          w.OrganizationID.String(),
		Name:                    w.Name,
		Description:             w.Description,
		Status:                  string(w.Status),
		ScheduleIntervalMinutes: w.ScheduleIntervalMinutes,
		Concurrency:             w.Concurrency,
		NotificationsEnabled:    w.NotificationsEnabled,
		NotifyOnSuccess:         w.NotifyOnSuccess,
		NotifyOnFailure:         w.NotifyOnFailure,
		NotifyOnCancel:          w.NotifyOnCancel,
		Timestamp:               w.UpdatedAt,
	})
}

func (w *Workflow) MarkDeleted() {
	w.Status = StatusDeleted
	w.UpdatedAt = time.Now().UTC()
	w.recordEvent(WorkflowDeleted{
		ID:             uuid.New().String(),
		WorkflowID:     w.ID.String(),
		OrganizationID: w.OrganizationID.String(),
		Timestamp:      w.UpdatedAt,
	})
}
