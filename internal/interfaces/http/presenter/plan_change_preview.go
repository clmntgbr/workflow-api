package presenter

import (
	"time"

	querysubscription "go-api/internal/application/query/subscription"
)

type ProrationPreviewLineResponse struct {
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Proration   bool   `json:"proration"`
}

type PlanChangePreviewResponse struct {
	RequiresCheckout bool                           `json:"requiresCheckout"`
	Currency         string                         `json:"currency"`
	AmountDue        int64                          `json:"amountDue"`
	Subtotal         int64                          `json:"subtotal"`
	Total            int64                          `json:"total"`
	ProrationDate    int64                          `json:"prorationDate,omitempty"`
	PeriodStart      time.Time                      `json:"periodStart"`
	PeriodEnd        time.Time                      `json:"periodEnd"`
	Lines            []ProrationPreviewLineResponse `json:"lines"`
	CurrentPlanID    string                         `json:"currentPlanId,omitempty"`
	CurrentPlanSlug  string                         `json:"currentPlanSlug,omitempty"`
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
			Description: line.Description,
			Amount:      line.Amount,
			Proration:   line.Proration,
		})
	}

	return PlanChangePreviewResponse{
		RequiresCheckout: preview.RequiresCheckout,
		Currency:         preview.Currency,
		AmountDue:        preview.AmountDue,
		Subtotal:         preview.Subtotal,
		Total:            preview.Total,
		ProrationDate:    preview.ProrationDate,
		PeriodStart:      preview.PeriodStart,
		PeriodEnd:        preview.PeriodEnd,
		Lines:            lines,
		CurrentPlanID:    preview.CurrentPlanID,
		CurrentPlanSlug:  preview.CurrentPlanSlug,
		TargetPlanID:     preview.TargetPlanID,
		TargetPlanSlug:   preview.TargetPlanSlug,
		TargetPlanName:   preview.TargetPlanName,
		TargetPlanPrice:  preview.TargetPlanPrice,
	}
}
