package presenter

import (
	"time"

	querysubscription "go-api/internal/application/query/subscription"
)

type ProrationPreviewLineResponse struct {
	Description *string `json:"description"`
	Amount      int64   `json:"amount"`
	Proration   bool    `json:"proration"`
}

type PlanChangePreviewResponse struct {
	RequiresCheckout bool                           `json:"requiresCheckout"`
	Currency         *string                        `json:"currency"`
	AmountDue        int64                          `json:"amountDue"`
	Subtotal         int64                          `json:"subtotal"`
	Total            int64                          `json:"total"`
	ProrationDate    *int64                         `json:"prorationDate"`
	PeriodStart      time.Time                      `json:"periodStart"`
	PeriodEnd        time.Time                      `json:"periodEnd"`
	Lines            []ProrationPreviewLineResponse `json:"lines"`
	CurrentPlanID    *string                        `json:"currentPlanId"`
	CurrentPlanSlug  *string                        `json:"currentPlanSlug"`
	TargetPlanID     string                         `json:"targetPlanId"`
	TargetPlanSlug   string                         `json:"targetPlanSlug"`
	TargetPlanName   string                         `json:"targetPlanName"`
	TargetPlanPrice  float64                        `json:"targetPlanPrice"`
}

func NewPlanChangePreviewResponse(preview *querysubscription.PlanChangePreview) PlanChangePreviewResponse {
	if preview == nil {
		return PlanChangePreviewResponse{Lines: []ProrationPreviewLineResponse{}}
	}

	lines := make([]ProrationPreviewLineResponse, 0, len(preview.Lines))
	for _, line := range preview.Lines {
		lines = append(lines, ProrationPreviewLineResponse{
			Description: optionalNonEmptyString(line.Description),
			Amount:      line.Amount,
			Proration:   line.Proration,
		})
	}

	var prorationDate *int64
	if preview.ProrationDate != 0 {
		prorationDate = &preview.ProrationDate
	}

	return PlanChangePreviewResponse{
		RequiresCheckout: preview.RequiresCheckout,
		Currency:         optionalNonEmptyString(preview.Currency),
		AmountDue:        preview.AmountDue,
		Subtotal:         preview.Subtotal,
		Total:            preview.Total,
		ProrationDate:    prorationDate,
		PeriodStart:      preview.PeriodStart,
		PeriodEnd:        preview.PeriodEnd,
		Lines:            lines,
		CurrentPlanID:    optionalNonEmptyString(preview.CurrentPlanID),
		CurrentPlanSlug:  optionalNonEmptyString(preview.CurrentPlanSlug),
		TargetPlanID:     preview.TargetPlanID,
		TargetPlanSlug:   preview.TargetPlanSlug,
		TargetPlanName:   preview.TargetPlanName,
		TargetPlanPrice:  preview.TargetPlanPrice,
	}
}
