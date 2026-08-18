package presenter

import (
	"time"

	domaininsight "go-api/internal/domain/insight"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type StepRunResponseSnapshot struct {
	Status int `json:"status"`
}

type StepRunResponse struct {
	ID               string                  `json:"id"`
	Name             string                  `json:"name"`
	URL              string                  `json:"url"`
	Method           string                  `json:"method"`
	ExecutionOrder   int                     `json:"executionOrder"`
	Status           string                  `json:"status"`
	Attempt          int                     `json:"attempt"`
	ResponseSnapshot *StepRunResponseSnapshot `json:"responseSnapshot"`
	Insights         []InsightResponse       `json:"insights,omitempty"`
	StartedAt        *time.Time              `json:"startedAt"`
	FinishedAt       *time.Time              `json:"finishedAt"`
	Error            *string                 `json:"error"`
}

func NewStepRunResponseFromView(
	view domainsteprun.StepRunView,
	insights []domaininsight.InsightView,
) StepRunResponse {
	return StepRunResponse{
		ID:               view.ID.String(),
		Name:             view.Name,
		URL:              view.URL,
		Method:           view.Method,
		ExecutionOrder:   view.ExecutionOrder,
		Status:           string(view.Status),
		Attempt:          view.Attempt,
		ResponseSnapshot: stepRunResponseSnapshot(view.ResponseSnapshot),
		Insights:         NewInsightListResponseFromViews(insights),
		StartedAt:        view.StartedAt,
		FinishedAt:       view.FinishedAt,
		Error:            optionalNonEmptyString(view.Error),
	}
}

func stepRunResponseSnapshot(snapshot *domainsteprun.ResponseSnapshot) *StepRunResponseSnapshot {
	if snapshot == nil {
		return nil
	}
	return &StepRunResponseSnapshot{Status: snapshot.Status}
}

func NewStepRunListResponseFromViews(
	views []domainsteprun.StepRunView,
	insightsByStepRunID map[uuid.UUID][]domaininsight.InsightView,
) []StepRunResponse {
	if len(views) == 0 {
		return nil
	}
	if insightsByStepRunID == nil {
		insightsByStepRunID = map[uuid.UUID][]domaininsight.InsightView{}
	}
	items := make([]StepRunResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewStepRunResponseFromView(view, insightsByStepRunID[view.ID]))
	}
	return items
}
