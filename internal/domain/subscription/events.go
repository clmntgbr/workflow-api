package subscription

import "time"

const (
	EventTypeSubscriptionCreated               = "subscription.created.v1"
	EventTypeSubscriptionUpdated               = "subscription.updated.v1"
	EventTypeSubscriptionPlanChanged           = "subscription.planChanged.v1"
	EventTypeSubscriptionRenewalUpcoming       = "subscription.renewalUpcoming.v1"
	EventTypeSubscriptionPaymentMethodExpiring = "subscription.paymentMethodExpiring.v1"
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

type SubscriptionPlanChanged struct {
	ID             string    `json:"eventId"`
	SubscriptionID string    `json:"subscriptionId"`
	PreviousPlanID string    `json:"previousPlanId"`
	PlanID         string    `json:"planId"`
	Status         string    `json:"status"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e SubscriptionPlanChanged) EventID() string       { return e.ID }
func (e SubscriptionPlanChanged) EventType() string     { return EventTypeSubscriptionPlanChanged }
func (e SubscriptionPlanChanged) AggregateID() string   { return e.SubscriptionID }
func (e SubscriptionPlanChanged) OccurredAt() time.Time { return e.Timestamp }

type SubscriptionRenewalUpcoming struct {
	ID             string    `json:"eventId"`
	SubscriptionID string    `json:"subscriptionId"`
	PlanID         string    `json:"planId"`
	AmountDue      int64     `json:"amountDue"`
	Currency       string    `json:"currency"`
	PeriodEnd      time.Time `json:"periodEnd"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e SubscriptionRenewalUpcoming) EventID() string       { return e.ID }
func (e SubscriptionRenewalUpcoming) EventType() string     { return EventTypeSubscriptionRenewalUpcoming }
func (e SubscriptionRenewalUpcoming) AggregateID() string   { return e.SubscriptionID }
func (e SubscriptionRenewalUpcoming) OccurredAt() time.Time { return e.Timestamp }

type SubscriptionPaymentMethodExpiring struct {
	ID             string    `json:"eventId"`
	SubscriptionID string    `json:"subscriptionId"`
	Brand          string    `json:"brand"`
	Last4          string    `json:"last4"`
	ExpMonth       int64     `json:"expMonth"`
	ExpYear        int64     `json:"expYear"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e SubscriptionPaymentMethodExpiring) EventID() string {
	return e.ID
}
func (e SubscriptionPaymentMethodExpiring) EventType() string {
	return EventTypeSubscriptionPaymentMethodExpiring
}
func (e SubscriptionPaymentMethodExpiring) AggregateID() string   { return e.SubscriptionID }
func (e SubscriptionPaymentMethodExpiring) OccurredAt() time.Time { return e.Timestamp }
