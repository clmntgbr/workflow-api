package presenter

import (
	"time"

	domaininsight "go-api/internal/domain/insight"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type WorkflowRunListResponse struct {
	ID          string     `json:"id"`
	Status      string     `json:"status"`
	TriggeredBy string     `json:"triggeredBy"`
	StartedAt   *time.Time `json:"startedAt"`
	FinishedAt  *time.Time `json:"finishedAt"`
	Error       *string    `json:"error"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type WorkflowRunDetailResponse struct {
	WorkflowRunListResponse
	Workflow *WorkflowDetailResponse `json:"workflow,omitempty"`
	StepRuns []StepRunResponse       `json:"stepRuns,omitempty"`
}

type WorkflowRunAnalyticsResponse struct {
	TotalRuns         int        `json:"totalRuns"`
	SuccessRate       float64    `json:"successRate"`
	SuccessCount      int        `json:"successCount"`
	FailureRate       float64    `json:"failureRate"`
	FailureCount      int        `json:"failureCount"`
	AverageDurationMS float64    `json:"averageDurationMs"`
	RunningCount      int        `json:"runningCount"`
	PendingCount      int        `json:"pendingCount"`
	CancelledCount    int        `json:"cancelledCount"`
	MinDurationMS     float64    `json:"minDurationMs"`
	MaxDurationMS     float64    `json:"maxDurationMs"`
	LastRunAt         *time.Time `json:"lastRunAt"`
}

func NewWorkflowRunListResponseFromView(view domainworkflowrun.WorkflowRunView) WorkflowRunListResponse {
	return WorkflowRunListResponse{
		ID:          view.ID.String(),
		Status:      string(view.Status),
		TriggeredBy: string(view.TriggeredBy),
		StartedAt:   view.StartedAt,
		FinishedAt:  view.FinishedAt,
		Error:       optionalNonEmptyString(view.Error),
		CreatedAt:   view.CreatedAt,
	}
}

func NewWorkflowRunDetailResponseFromEntity(
	run domainworkflowrun.WorkflowRun,
	_ uuid.UUID,
) WorkflowRunDetailResponse {
	return WorkflowRunDetailResponse{
		WorkflowRunListResponse: WorkflowRunListResponse{
			ID:          run.ID.String(),
			Status:      string(run.Status),
			TriggeredBy: string(run.TriggeredBy),
			StartedAt:   run.StartedAt,
			FinishedAt:  run.FinishedAt,
			Error:       optionalNonEmptyString(run.Error),
			CreatedAt:   run.CreatedAt,
		},
	}
}

func NewWorkflowRunDetailResponseFromView(
	view domainworkflowrun.WorkflowRunView,
	stepRuns []domainsteprun.StepRunView,
	insightsByStepRunID map[uuid.UUID][]domaininsight.InsightView,
) WorkflowRunDetailResponse {
	return NewWorkflowRunDetailResponseFromViewWithRelations(view, nil, stepRuns, nil, insightsByStepRunID)
}

func NewWorkflowRunDetailResponseFromViewWithRelations(
	view domainworkflowrun.WorkflowRunView,
	workflow *domainworkflow.WorkflowView,
	stepRuns []domainsteprun.StepRunView,
	stepsByID map[uuid.UUID]domainstep.StepView,
	insightsByStepRunID map[uuid.UUID][]domaininsight.InsightView,
) WorkflowRunDetailResponse {
	resp := WorkflowRunDetailResponse{
		WorkflowRunListResponse: NewWorkflowRunListResponseFromView(view),
		StepRuns: NewStepRunListResponseFromViewsWithSteps(
			stepRuns,
			stepsByID,
			insightsByStepRunID,
		),
	}
	if workflow != nil {
		detail := NewWorkflowDetailResponseFromView(*workflow)
		resp.Workflow = &detail
	}
	return resp
}

func NewWorkflowRunListWithStepRunsFromViews(
	views []domainworkflowrun.WorkflowRunView,
	stepRunsByWorkflowRunID map[uuid.UUID][]domainsteprun.StepRunView,
	insightsByStepRunID map[uuid.UUID][]domaininsight.InsightView,
) []WorkflowRunDetailResponse {
	items := make([]WorkflowRunDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewWorkflowRunDetailResponseFromView(
			view,
			stepRunsByWorkflowRunID[view.ID],
			insightsByStepRunID,
		))
	}
	return items
}

func NewWorkflowRunAnalyticsResponse(
	stats domainworkflowrun.WorkflowRunAnalytics,
) WorkflowRunAnalyticsResponse {
	return WorkflowRunAnalyticsResponse{
		TotalRuns:         int(stats.TotalRuns),
		SuccessRate:       stats.SuccessRate,
		SuccessCount:      int(stats.SuccessCount),
		FailureRate:       stats.FailureRate,
		FailureCount:      int(stats.FailureCount),
		AverageDurationMS: stats.AverageDurationMS,
		RunningCount:      int(stats.RunningCount),
		PendingCount:      int(stats.PendingCount),
		CancelledCount:    int(stats.CancelledCount),
		MinDurationMS:     stats.MinDurationMS,
		MaxDurationMS:     stats.MaxDurationMS,
		LastRunAt:         stats.LastRunAt,
	}
}
