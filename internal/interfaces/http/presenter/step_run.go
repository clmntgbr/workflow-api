package presenter

import (
	"time"

	domaininsight "go-api/internal/domain/insight"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type StepRunResponseSnapshot struct {
	Status int `json:"status"`
}

type StepRunResponse struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	URL              string                    `json:"url"`
	Method           string                    `json:"method"`
	ExecutionOrder   int                       `json:"executionOrder"`
	Status           string                    `json:"status"`
	Attempt          int                       `json:"attempt"`
	ResponseSnapshot *StepRunResponseSnapshot  `json:"responseSnapshot"`
	AssertionsResult []AssertionResultResponse `json:"assertionsResult"`
	Insights         []InsightResponse         `json:"insights"`
	Step             *StepDetailResponse       `json:"step"`
	StartedAt        *time.Time                `json:"startedAt"`
	FinishedAt       *time.Time                `json:"finishedAt"`
	ResumeAt         *time.Time                `json:"resumeAt,omitempty"`
	MatchedBranch    *bool                     `json:"matchedBranch,omitempty"`
	Error            *string                   `json:"error"`
}

func NewStepRunResponseFromView(
	view domainsteprun.StepRunView,
	insights []domaininsight.InsightView,
) StepRunResponse {
	return NewStepRunResponseFromViewWithStep(view, nil, insights)
}

func NewStepRunResponseFromViewWithStep(
	view domainsteprun.StepRunView,
	step *domainstep.StepView,
	insights []domaininsight.InsightView,
) StepRunResponse {
	resp := StepRunResponse{
		ID:               view.ID.String(),
		Name:             view.Name,
		URL:              view.URL,
		Method:           view.Method,
		ExecutionOrder:   view.ExecutionOrder,
		Status:           string(view.Status),
		Attempt:          view.Attempt,
		ResponseSnapshot: stepRunResponseSnapshot(view.ResponseSnapshot),
		AssertionsResult: NewAssertionResultListResponse(
			view.AssertionsResult,
			view.Assertions,
			view.StepID,
			view.WorkflowID,
		),
		Insights:   NewInsightListResponseFromViews(insights),
		StartedAt:  view.StartedAt,
		FinishedAt: view.FinishedAt,
		ResumeAt:   view.ResumeAt,
		MatchedBranch: view.MatchedBranch,
		Error:      optionalNonEmptyString(view.Error),
	}
	if step != nil {
		detail := NewStepDetailResponseFromView(*step)
		resp.Step = &detail
	}
	return resp
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
	return NewStepRunListResponseFromViewsWithSteps(views, nil, insightsByStepRunID)
}

func NewStepRunListResponseFromViewsWithSteps(
	views []domainsteprun.StepRunView,
	stepsByID map[uuid.UUID]domainstep.StepView,
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
		var step *domainstep.StepView
		if stepsByID != nil {
			if found, ok := stepsByID[view.StepID]; ok {
				stepCopy := found
				step = &stepCopy
			}
		}
		items = append(items, NewStepRunResponseFromViewWithStep(
			view,
			step,
			insightsByStepRunID[view.ID],
		))
	}
	return items
}
