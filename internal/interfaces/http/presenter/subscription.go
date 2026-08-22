package presenter

import (
	"time"

	domainsubscription "go-api/internal/domain/subscription"
)

type SubscriptionResponse struct {
	ID                   string        `json:"id"`
	Status               string        `json:"status"`
	StripeCustomerID     string        `json:"stripeCustomerId"`
	StripeSubscriptionID string        `json:"stripeSubscriptionId"`
	StartDate            time.Time     `json:"startDate"`
	EndDate              time.Time     `json:"endDate"`
	CancelAtPeriodEnd    bool          `json:"cancelAtPeriodEnd"`
	QuotaPeriodStart     time.Time     `json:"quotaPeriodStart"`
	Plan                 *PlanResponse `json:"plan"`
	CreatedAt            time.Time     `json:"createdAt"`
	UpdatedAt            time.Time     `json:"updatedAt"`
}

func NewSubscriptionResponse(view *domainsubscription.SubscriptionView) SubscriptionResponse {
	var plan *PlanResponse
	if view.Plan != nil {
		p := NewPlanResponseFromView(*view.Plan)
		plan = &p
	}

	return SubscriptionResponse{
		ID:                   view.ID.String(),
		Status:               string(view.Status),
		StripeCustomerID:     view.StripeCustomerID,
		StripeSubscriptionID: view.StripeSubscriptionID,
		StartDate:            view.StartDate,
		EndDate:              view.EndDate,
		CancelAtPeriodEnd:    view.CancelAtPeriodEnd,
		QuotaPeriodStart:     view.QuotaPeriodStart,
		Plan:                 plan,
		CreatedAt:            view.CreatedAt,
		UpdatedAt:            view.UpdatedAt,
	}
}
