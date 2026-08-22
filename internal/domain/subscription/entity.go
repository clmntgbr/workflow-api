package subscription

import (
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Subscription struct {
	ID uuid.UUID

	PlanID uuid.UUID

	StripeCustomerID     string
	StripeSubscriptionID string

	Status    Status
	StartDate time.Time
	EndDate   time.Time

	CancelAtPeriodEnd bool
	QuotaPeriodStart  time.Time

	CreatedAt time.Time
	UpdatedAt time.Time

	events []event.DomainEvent
}

func NewSubscription(planID uuid.UUID, status Status, startDate, endDate time.Time) *Subscription {
	now := time.Now().UTC()
	if startDate.IsZero() {
		startDate = now
	}
	if endDate.IsZero() {
		endDate = startDate
	}
	s := &Subscription{
		ID:                uuid.New(),
		PlanID:            planID,
		Status:            status,
		StartDate:         startDate.UTC(),
		EndDate:           endDate.UTC(),
		CancelAtPeriodEnd: false,
		QuotaPeriodStart:  startDate.UTC(),
		CreatedAt:         now,
		UpdatedAt:         now,
	}
	s.recordEvent(SubscriptionCreated{
		ID:                   uuid.New().String(),
		SubscriptionID:       s.ID.String(),
		PlanID:               s.PlanID.String(),
		StripeCustomerID:     s.StripeCustomerID,
		StripeSubscriptionID: s.StripeSubscriptionID,
		Status:               string(s.Status),
		StartDate:            s.StartDate,
		EndDate:              s.EndDate,
		CancelAtPeriodEnd:    s.CancelAtPeriodEnd,
		QuotaPeriodStart:     s.QuotaPeriodStart,
		Timestamp:            now,
	})
	return s
}

func NewFreeSubscription(planID uuid.UUID) *Subscription {
	now := time.Now().UTC()
	return NewSubscription(planID, StatusActive, now, now.AddDate(100, 0, 0))
}

func (s *Subscription) PullEvents() []event.DomainEvent {
	events := s.events
	s.events = nil
	return events
}

func (s *Subscription) recordEvent(e event.DomainEvent) {
	s.events = append(s.events, e)
}

func (s *Subscription) ApplyUpdate(
	planID uuid.UUID,
	status Status,
	stripeCustomerID string,
	stripeSubscriptionID string,
	startDate time.Time,
	endDate time.Time,
	cancelAtPeriodEnd bool,
	quotaPeriodStart time.Time,
) {
	s.PlanID = planID
	s.Status = status
	s.StripeCustomerID = stripeCustomerID
	s.StripeSubscriptionID = stripeSubscriptionID
	s.StartDate = startDate.UTC()
	s.EndDate = endDate.UTC()
	s.CancelAtPeriodEnd = cancelAtPeriodEnd
	if !quotaPeriodStart.IsZero() {
		s.QuotaPeriodStart = quotaPeriodStart.UTC()
	}
	s.UpdatedAt = time.Now().UTC()
	s.recordEvent(SubscriptionUpdated{
		ID:                   uuid.New().String(),
		SubscriptionID:       s.ID.String(),
		PlanID:               s.PlanID.String(),
		StripeCustomerID:     s.StripeCustomerID,
		StripeSubscriptionID: s.StripeSubscriptionID,
		Status:               string(s.Status),
		StartDate:            s.StartDate,
		EndDate:              s.EndDate,
		CancelAtPeriodEnd:    s.CancelAtPeriodEnd,
		QuotaPeriodStart:     s.QuotaPeriodStart,
		Timestamp:            s.UpdatedAt,
	})
}

func (s *Subscription) DowngradeToFree(freePlanID uuid.UUID) {
	now := time.Now().UTC()
	s.ApplyUpdate(
		freePlanID,
		StatusActive,
		s.StripeCustomerID,
		"",
		now,
		now.AddDate(100, 0, 0),
		false,
		s.QuotaPeriodStart,
	)
}
