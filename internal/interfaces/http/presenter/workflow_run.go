package presenter

import (
	"time"

	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type WorkflowRunDetailResponse struct {
	ID                string            `json:"id"`
	WorkflowID        string            `json:"workflowId"`
	OrganizationID    string            `json:"organizationId"`
	Status            string            `json:"status"`
	TriggeredBy       string            `json:"triggeredBy"`
	TriggeredByUserID *string           `json:"triggeredByUserId"`
	Context           map[string]any    `json:"context"`
	StartedAt         *time.Time        `json:"startedAt"`
	FinishedAt        *time.Time        `json:"finishedAt"`
	Error             *string           `json:"error"`
	CreatedAt         time.Time         `json:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt"`
	StepRuns          []StepRunResponse `json:"stepRuns,omitempty"`
}

func NewWorkflowRunDetailResponseFromEntity(
	run domainworkflowrun.WorkflowRun,
	organizationID uuid.UUID,
) WorkflowRunDetailResponse {
	return WorkflowRunDetailResponse{
		ID:                run.ID.String(),
		WorkflowID:        run.WorkflowID.String(),
		OrganizationID:    organizationID.String(),
		Status:            string(run.Status),
		TriggeredBy:       string(run.TriggeredBy),
		TriggeredByUserID: optionalUUID(run.TriggeredByUserID),
		Context:           contextOrEmpty(run.Context),
		StartedAt:         run.StartedAt,
		FinishedAt:        run.FinishedAt,
		Error:             optionalNonEmptyString(run.Error),
		CreatedAt:         run.CreatedAt,
		UpdatedAt:         run.UpdatedAt,
	}
}

func NewWorkflowRunDetailResponseFromView(
	view domainworkflowrun.WorkflowRunView,
	stepRuns []domainsteprun.StepRunView,
) WorkflowRunDetailResponse {
	return WorkflowRunDetailResponse{
		ID:                view.ID.String(),
		WorkflowID:        view.WorkflowID.String(),
		OrganizationID:    view.OrganizationID.String(),
		Status:            string(view.Status),
		TriggeredBy:       string(view.TriggeredBy),
		TriggeredByUserID: optionalUUID(view.TriggeredByUserID),
		Context:           contextOrEmpty(view.Context),
		StartedAt:         view.StartedAt,
		FinishedAt:        view.FinishedAt,
		Error:             optionalNonEmptyString(view.Error),
		CreatedAt:         view.CreatedAt,
		UpdatedAt:         view.UpdatedAt,
		StepRuns:          NewStepRunListResponseFromViews(stepRuns),
	}
}

func NewWorkflowRunListResponseFromViews(views []domainworkflowrun.WorkflowRunView) []WorkflowRunDetailResponse {
	items := make([]WorkflowRunDetailResponse, 0, len(views))
	for _, view := range views {
		item := NewWorkflowRunDetailResponseFromView(view, nil)
		item.StepRuns = nil
		items = append(items, item)
	}
	return items
}

func NewWorkflowRunListWithStepRunsFromViews(
	views []domainworkflowrun.WorkflowRunView,
	stepRunsByWorkflowRunID map[uuid.UUID][]domainsteprun.StepRunView,
) []WorkflowRunDetailResponse {
	items := make([]WorkflowRunDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewWorkflowRunDetailResponseFromView(view, stepRunsByWorkflowRunID[view.ID]))
	}
	return items
}

func optionalUUID(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	value := id.String()
	return &value
}

func contextOrEmpty(ctx map[string]any) map[string]any {
	if ctx == nil {
		return map[string]any{}
	}
	return ctx
}
