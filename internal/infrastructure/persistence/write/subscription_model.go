package write

import (
	"time"

	domainsubscription "go-api/internal/domain/subscription"

	"github.com/google/uuid"
)

type SubscriptionModel struct {
	ID                   uuid.UUID `gorm:"column:id;primaryKey"`
	PlanID               uuid.UUID `gorm:"column:plan_id"`
	StripeCustomerID     string    `gorm:"column:stripe_customer_id"`
	StripeSubscriptionID string    `gorm:"column:stripe_subscription_id"`
	Status               string    `gorm:"column:status"`
	StartDate            time.Time `gorm:"column:start_date"`
	EndDate              time.Time `gorm:"column:end_date"`
	CancelAtPeriodEnd    bool      `gorm:"column:cancel_at_period_end"`
	QuotaPeriodStart     time.Time `gorm:"column:quota_period_start"`
	CreatedAt            time.Time `gorm:"column:created_at"`
	UpdatedAt            time.Time `gorm:"column:updated_at"`
}

func (SubscriptionModel) TableName() string { return "subscriptions" }

func subscriptionModelFromDomain(s *domainsubscription.Subscription) *SubscriptionModel {
	return &SubscriptionModel{
		ID:                   s.ID,
		PlanID:               s.PlanID,
		StripeCustomerID:     s.StripeCustomerID,
		StripeSubscriptionID: s.StripeSubscriptionID,
		Status:               string(s.Status),
		StartDate:            s.StartDate,
		EndDate:              s.EndDate,
		CancelAtPeriodEnd:    s.CancelAtPeriodEnd,
		QuotaPeriodStart:     s.QuotaPeriodStart,
		CreatedAt:            s.CreatedAt,
		UpdatedAt:            s.UpdatedAt,
	}
}

func subscriptionDomainFromModel(model *SubscriptionModel) *domainsubscription.Subscription {
	return &domainsubscription.Subscription{
		ID:                   model.ID,
		PlanID:               model.PlanID,
		StripeCustomerID:     model.StripeCustomerID,
		StripeSubscriptionID: model.StripeSubscriptionID,
		Status:               domainsubscription.Status(model.Status),
		StartDate:            model.StartDate,
		EndDate:              model.EndDate,
		CancelAtPeriodEnd:    model.CancelAtPeriodEnd,
		QuotaPeriodStart:     model.QuotaPeriodStart,
		CreatedAt:            model.CreatedAt,
		UpdatedAt:            model.UpdatedAt,
	}
}
