package user

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	ClerkID   string
	FirstName string
	LastName  string
	Banned    bool
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewUser(clerkID, firstName, lastName, email string, banned bool) *User {
	now := time.Now().UTC()
	u := &User{
		ID:        uuid.New(),
		ClerkID:   clerkID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Banned:    banned,
		CreatedAt: now,
		UpdatedAt: now,
	}
	u.recordEvent(UserCreated{
		ID:        uuid.New().String(),
		UserID:    u.ID.String(),
		ClerkID:   clerkID,
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
		Banned:    banned,
		Timestamp: now,
	})
	return u
}

func (u *User) PullEvents() []event.DomainEvent {
	events := u.events
	u.events = nil
	return events
}

func (u *User) recordEvent(e event.DomainEvent) {
	u.events = append(u.events, e)
}

func (u *User) ApplyUpdate(firstName, lastName, email string, banned bool) {
	u.FirstName = firstName
	u.LastName = lastName
	u.Banned = banned
	if email != "" {
		u.Email = email
	}
	u.UpdatedAt = time.Now().UTC()
	u.recordEvent(UserUpdated{
		ID:        uuid.New().String(),
		UserID:    u.ID.String(),
		ClerkID:   u.ClerkID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Banned:    u.Banned,
		Timestamp: u.UpdatedAt,
	})
}

func (u *User) MarkDeleted() {
	u.recordEvent(UserDeleted{
		ID:        uuid.New().String(),
		UserID:    u.ID.String(),
		ClerkID:   u.ClerkID,
		Timestamp: time.Now().UTC(),
	})
}
