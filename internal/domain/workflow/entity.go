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

	ScheduleType          ScheduleType
	ScheduleIntervalValue int
	ScheduleIntervalUnit  ScheduleUnit
	ScheduleAt            *time.Time
	ScheduleTimezone      string
	NextRunAt             *time.Time

	Concurrency int

	NotificationsEnabled bool
	NotifyOnSuccess      bool
	NotifyOnFailure      bool
	NotifyOnCancel       bool

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

type NewWorkflowParams struct {
	Name                  string
	Description           string
	OrganizationID        uuid.UUID
	ScheduleType          ScheduleType
	ScheduleIntervalValue int
	ScheduleIntervalUnit  ScheduleUnit
	ScheduleAt            *time.Time
	ScheduleTimezone      string
	Concurrency           int
	NotificationsEnabled  bool
	NotifyOnSuccess       bool
	NotifyOnFailure       bool
	NotifyOnCancel        bool
}

func NewWorkflow(p NewWorkflowParams) (*Workflow, error) {
	now := time.Now().UTC()
	concurrency := p.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	scheduleType := p.ScheduleType
	if scheduleType == "" {
		scheduleType = ScheduleTypeNone
	}
	timezone := p.ScheduleTimezone
	if timezone == "" {
		timezone = "UTC"
	}

	w := &Workflow{
		ID:                    uuid.New(),
		Name:                  p.Name,
		Description:           p.Description,
		Status:                StatusInactive,
		OrganizationID:        p.OrganizationID,
		ScheduleType:          scheduleType,
		ScheduleIntervalValue: p.ScheduleIntervalValue,
		ScheduleIntervalUnit:  p.ScheduleIntervalUnit,
		ScheduleAt:            p.ScheduleAt,
		ScheduleTimezone:      timezone,
		Concurrency:           concurrency,
		NotificationsEnabled:  p.NotificationsEnabled,
		NotifyOnSuccess:       p.NotifyOnSuccess,
		NotifyOnFailure:       p.NotifyOnFailure,
		NotifyOnCancel:        p.NotifyOnCancel,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := w.validateSchedule(); err != nil {
		return nil, err
	}
	w.RecalculateNextRunAt(now)
	w.recordEvent(w.createdEvent(now))
	return w, nil
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
	Name                  string
	Description           string
	Status                Status
	ScheduleType          ScheduleType
	ScheduleIntervalValue int
	ScheduleIntervalUnit  ScheduleUnit
	ScheduleAt            *time.Time
	ScheduleTimezone      string
	Concurrency           int
	NotificationsEnabled  bool
	NotifyOnSuccess       bool
	NotifyOnFailure       bool
	NotifyOnCancel        bool
}

func (w *Workflow) ApplyUpdate(p UpdateWorkflowParams) error {
	concurrency := p.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	timezone := p.ScheduleTimezone
	if timezone == "" {
		timezone = "UTC"
	}
	scheduleType := p.ScheduleType
	if scheduleType == "" {
		scheduleType = ScheduleTypeNone
	}

	w.Name = p.Name
	w.Description = p.Description
	w.Status = p.Status
	w.ScheduleType = scheduleType
	w.ScheduleIntervalValue = p.ScheduleIntervalValue
	w.ScheduleIntervalUnit = p.ScheduleIntervalUnit
	w.ScheduleAt = p.ScheduleAt
	w.ScheduleTimezone = timezone
	w.Concurrency = concurrency
	w.NotificationsEnabled = p.NotificationsEnabled
	w.NotifyOnSuccess = p.NotifyOnSuccess
	w.NotifyOnFailure = p.NotifyOnFailure
	w.NotifyOnCancel = p.NotifyOnCancel
	w.UpdatedAt = time.Now().UTC()

	if err := w.validateSchedule(); err != nil {
		return err
	}
	w.RecalculateNextRunAt(w.UpdatedAt)
	w.recordEvent(w.updatedEvent(w.UpdatedAt))
	return nil
}

func (w *Workflow) Activate() error {
	if w.Status == StatusDeleted || w.Status == StatusCanceled {
		return ErrInvalidStatusTransition
	}
	if w.Status == StatusActive {
		return nil
	}
	now := time.Now().UTC()
	w.Status = StatusActive
	w.UpdatedAt = now
	w.RecalculateNextRunAt(now)
	w.recordEvent(w.updatedEvent(now))
	return nil
}

func (w *Workflow) Deactivate() error {
	if w.Status == StatusDeleted || w.Status == StatusCanceled {
		return ErrInvalidStatusTransition
	}
	if w.Status == StatusInactive {
		return nil
	}
	now := time.Now().UTC()
	w.Status = StatusInactive
	w.NextRunAt = nil
	w.UpdatedAt = now
	w.recordEvent(w.updatedEvent(now))
	return nil
}

func (w *Workflow) MarkDeleted() {
	w.Status = StatusDeleted
	w.ScheduleType = ScheduleTypeNone
	w.NextRunAt = nil
	w.UpdatedAt = time.Now().UTC()
	w.recordEvent(WorkflowDeleted{
		ID:             uuid.New().String(),
		WorkflowID:     w.ID.String(),
		OrganizationID: w.OrganizationID.String(),
		Timestamp:      w.UpdatedAt,
	})
}

func (w *Workflow) createdEvent(at time.Time) WorkflowCreated {
	return WorkflowCreated{
		ID:                    uuid.New().String(),
		WorkflowID:            w.ID.String(),
		OrganizationID:        w.OrganizationID.String(),
		Name:                  w.Name,
		Description:           w.Description,
		Status:                string(w.Status),
		ScheduleType:          string(w.ScheduleType),
		ScheduleIntervalValue: w.ScheduleIntervalValue,
		ScheduleIntervalUnit:  string(w.ScheduleIntervalUnit),
		ScheduleAt:            w.ScheduleAt,
		ScheduleTimezone:      w.ScheduleTimezone,
		NextRunAt:             w.NextRunAt,
		Concurrency:           w.Concurrency,
		NotificationsEnabled:  w.NotificationsEnabled,
		NotifyOnSuccess:       w.NotifyOnSuccess,
		NotifyOnFailure:       w.NotifyOnFailure,
		NotifyOnCancel:        w.NotifyOnCancel,
		Timestamp:             at,
	}
}

func (w *Workflow) updatedEvent(at time.Time) WorkflowUpdated {
	return WorkflowUpdated{
		ID:                    uuid.New().String(),
		WorkflowID:            w.ID.String(),
		OrganizationID:        w.OrganizationID.String(),
		Name:                  w.Name,
		Description:           w.Description,
		Status:                string(w.Status),
		ScheduleType:          string(w.ScheduleType),
		ScheduleIntervalValue: w.ScheduleIntervalValue,
		ScheduleIntervalUnit:  string(w.ScheduleIntervalUnit),
		ScheduleAt:            w.ScheduleAt,
		ScheduleTimezone:      w.ScheduleTimezone,
		NextRunAt:             w.NextRunAt,
		Concurrency:           w.Concurrency,
		NotificationsEnabled:  w.NotificationsEnabled,
		NotifyOnSuccess:       w.NotifyOnSuccess,
		NotifyOnFailure:       w.NotifyOnFailure,
		NotifyOnCancel:        w.NotifyOnCancel,
		Timestamp:             at,
	}
}
