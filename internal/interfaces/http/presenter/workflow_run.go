package presenter

import (
	"time"

	domaininsight "go-api/internal/domain/insight"
	domainsteprun "go-api/internal/domain/steprun"
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
	StepRuns []StepRunResponse `json:"stepRuns,omitempty"`
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
	return WorkflowRunDetailResponse{
		WorkflowRunListResponse: NewWorkflowRunListResponseFromView(view),
		StepRuns:                NewStepRunListResponseFromViews(stepRuns, insightsByStepRunID),
	}
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
