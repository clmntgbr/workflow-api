package project

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Project struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	MemberIDs []uuid.UUID

	events []event.DomainEvent
}

func NewProject(name string, createdByUserID uuid.UUID) *Project {
	now := time.Now().UTC()
	org := &Project{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		MemberIDs: nil,
	}
	org.recordEvent(ProjectCreated{
		ID:              uuid.New().String(),
		ProjectID:  org.ID.String(),
		Name:            org.Name,
		CreatedByUserID: createdByUserID.String(),
		Timestamp:       now,
	})
	return org
}

func (o *Project) PullEvents() []event.DomainEvent {
	events := o.events
	o.events = nil
	return events
}

func (o *Project) recordEvent(e event.DomainEvent) {
	o.events = append(o.events, e)
}

func (o *Project) ApplyUpdate(name string) {
	o.Name = name
	o.UpdatedAt = time.Now().UTC()
	o.recordEvent(ProjectUpdated{
		ID:             uuid.New().String(),
		ProjectID: o.ID.String(),
		Name:           o.Name,
		MemberIDs:      uuidStrings(o.MemberIDs),
		Timestamp:      o.UpdatedAt,
	})
}

func (o *Project) MarkDeleted() {
	o.recordEvent(ProjectDeleted{
		ID:             uuid.New().String(),
		ProjectID: o.ID.String(),
		MemberIDs:      uuidStrings(o.MemberIDs),
		Timestamp:      time.Now().UTC(),
	})
}

func uuidStrings(ids []uuid.UUID) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return out
}

func (o *Project) AddMember(userID uuid.UUID) bool {
	for _, id := range o.MemberIDs {
		if id == userID {
			return false
		}
	}
	o.MemberIDs = append(o.MemberIDs, userID)
	o.UpdatedAt = time.Now().UTC()
	o.recordEvent(ProjectMemberAdded{
		ID:             uuid.New().String(),
		ProjectID: o.ID.String(),
		UserID:         userID.String(),
		Timestamp:      o.UpdatedAt,
	})
	return true
}

func (o *Project) RemoveMember(userID uuid.UUID) bool {
	idx := -1
	for i, id := range o.MemberIDs {
		if id == userID {
			idx = i
			break
		}
	}
	if idx < 0 {
		return false
	}
	o.MemberIDs = append(o.MemberIDs[:idx], o.MemberIDs[idx+1:]...)
	o.UpdatedAt = time.Now().UTC()
	o.recordEvent(ProjectMemberRemoved{
		ID:             uuid.New().String(),
		ProjectID: o.ID.String(),
		UserID:         userID.String(),
		Timestamp:      o.UpdatedAt,
	})
	return true
}
