package subscription

import "time"

const (
	EventTypeSubscriptionCreated = "subscription.created.v1"
	EventTypeSubscriptionUpdated = "subscription.updated.v1"
)

type SubscriptionCreated struct {
	ID                   string    `json:"eventId"`
	SubscriptionID       string    `json:"subscriptionId"`
	PlanID               string    `json:"planId"`
	StripeCustomerID     string    `json:"stripeCustomerId"`
	StripeSubscriptionID string    `json:"stripeSubscriptionId"`
	Status               string    `json:"status"`
	StartDate            time.Time `json:"startDate"`
	EndDate              time.Time `json:"endDate"`
	CancelAtPeriodEnd    bool      `json:"cancelAtPeriodEnd"`
	QuotaPeriodStart     time.Time `json:"quotaPeriodStart"`
	Timestamp            time.Time `json:"timestamp"`
}

func (e SubscriptionCreated) EventID() string       { return e.ID }
func (e SubscriptionCreated) EventType() string     { return EventTypeSubscriptionCreated }
func (e SubscriptionCreated) AggregateID() string   { return e.SubscriptionID }
func (e SubscriptionCreated) OccurredAt() time.Time { return e.Timestamp }

type SubscriptionUpdated struct {
	ID                   string    `json:"eventId"`
	SubscriptionID       string    `json:"subscriptionId"`
	PlanID               string    `json:"planId"`
	StripeCustomerID     string    `json:"stripeCustomerId"`
	StripeSubscriptionID string    `json:"stripeSubscriptionId"`
	Status               string    `json:"status"`
	StartDate            time.Time `json:"startDate"`
	EndDate              time.Time `json:"endDate"`
	CancelAtPeriodEnd    bool      `json:"cancelAtPeriodEnd"`
	QuotaPeriodStart     time.Time `json:"quotaPeriodStart"`
	Timestamp            time.Time `json:"timestamp"`
}

func (e SubscriptionUpdated) EventID() string       { return e.ID }
func (e SubscriptionUpdated) EventType() string     { return EventTypeSubscriptionUpdated }
func (e SubscriptionUpdated) AggregateID() string   { return e.SubscriptionID }
func (e SubscriptionUpdated) OccurredAt() time.Time { return e.Timestamp }
