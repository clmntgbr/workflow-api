package organization

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

// Organization is the aggregate root for an organization and its memberships.
// Members are referenced by user ID only — User remains a separate aggregate.
// Whether an organization is "active" for a user is derived from User.ActiveOrganizationID.
type Organization struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
	MemberIDs []uuid.UUID

	events []event.DomainEvent
}

func NewOrganization(name string, createdByUserID uuid.UUID) *Organization {
	now := time.Now().UTC()
	org := &Organization{
		ID:        uuid.New(),
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
		MemberIDs: nil,
	}
	org.recordEvent(OrganizationCreated{
		ID:              uuid.New().String(),
		OrganizationID:  org.ID.String(),
		Name:            org.Name,
		CreatedByUserID: createdByUserID.String(),
		Timestamp:       now,
	})
	return org
}

func (o *Organization) PullEvents() []event.DomainEvent {
	events := o.events
	o.events = nil
	return events
}

func (o *Organization) recordEvent(e event.DomainEvent) {
	o.events = append(o.events, e)
}

func (o *Organization) ApplyUpdate(name string) {
	o.Name = name
	o.UpdatedAt = time.Now().UTC()
	o.recordEvent(OrganizationUpdated{
		ID:             uuid.New().String(),
		OrganizationID: o.ID.String(),
		Name:           o.Name,
		Timestamp:      o.UpdatedAt,
	})
}

func (o *Organization) MarkDeleted() {
	o.recordEvent(OrganizationDeleted{
		ID:             uuid.New().String(),
		OrganizationID: o.ID.String(),
		Timestamp:      time.Now().UTC(),
	})
}

func (o *Organization) AddMember(userID uuid.UUID) bool {
	for _, id := range o.MemberIDs {
		if id == userID {
			return false
		}
	}
	o.MemberIDs = append(o.MemberIDs, userID)
	o.UpdatedAt = time.Now().UTC()
	o.recordEvent(OrganizationMemberAdded{
		ID:             uuid.New().String(),
		OrganizationID: o.ID.String(),
		UserID:         userID.String(),
		Timestamp:      o.UpdatedAt,
	})
	return true
}

func (o *Organization) RemoveMember(userID uuid.UUID) bool {
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
	o.recordEvent(OrganizationMemberRemoved{
		ID:             uuid.New().String(),
		OrganizationID: o.ID.String(),
		UserID:         userID.String(),
		Timestamp:      o.UpdatedAt,
	})
	return true
}
