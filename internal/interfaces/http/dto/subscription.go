package dto

type PreviewSubscriptionRequest struct {
	PlanID string `json:"planId" validate:"required,uuid"`
}

type CreateSubscriptionRequest struct {
	PlanID        string `json:"planId" validate:"required,uuid"`
	ProrationDate *int64 `json:"prorationDate" validate:"omitempty,gte=0"`
}
