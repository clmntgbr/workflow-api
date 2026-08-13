package presenter

import (
	"time"

	domainsteprun "go-api/internal/domain/steprun"
)

type StepRunResponse struct {
	ID               string                         `json:"id"`
	WorkflowRunID    string                         `json:"workflowRunId"`
	StepID           string                         `json:"stepId"`
	WorkflowID       string                         `json:"workflowId"`
	EndpointID       string                         `json:"endpointId"`
	OrganizationID   string                         `json:"organizationId"`
	Name             string                         `json:"name"`
	Description      *string                        `json:"description"`
	URL              string                         `json:"url"`
	Method           string                         `json:"method"`
	Headers          map[string]string              `json:"headers"`
	Query            map[string]string              `json:"query"`
	Body             map[string]any                 `json:"body"`
	Timeout          int                            `json:"timeout"`
	RetryOnFailure   bool                           `json:"retryOnFailure"`
	RetryCount       int                            `json:"retryCount"`
	RetryDelay       int                            `json:"retryDelay"`
	Index            string                         `json:"index"`
	ExecutionOrder   int                            `json:"executionOrder"`
	TreeIndex        int                            `json:"treeIndex"`
	Position         StepPositionResponse           `json:"position"`
	Status           string                         `json:"status"`
	Attempt          int                            `json:"attempt"`
	ResponseSnapshot *domainsteprun.ResponseSnapshot `json:"responseSnapshot"`
	StartedAt        *time.Time                     `json:"startedAt"`
	FinishedAt       *time.Time                     `json:"finishedAt"`
	Error            *string                        `json:"error"`
	CreatedAt        time.Time                      `json:"createdAt"`
	UpdatedAt        time.Time                      `json:"updatedAt"`
}

func NewStepRunResponseFromView(view domainsteprun.StepRunView) StepRunResponse {
	return StepRunResponse{
		ID:             view.ID.String(),
		WorkflowRunID:  view.WorkflowRunID.String(),
		StepID:         view.StepID.String(),
		WorkflowID:     view.WorkflowID.String(),
		EndpointID:     view.EndpointID.String(),
		OrganizationID: view.OrganizationID.String(),
		Name:           view.Name,
		Description:    optionalNonEmptyString(view.Description),
		URL:            view.URL,
		Method:         view.Method,
		Headers:        view.Headers,
		Query:          view.Query,
		Body:           view.Body,
		Timeout:        view.Timeout,
		RetryOnFailure: view.RetryOnFailure,
		RetryCount:     view.RetryCount,
		RetryDelay:     view.RetryDelay,
		Index:          view.Index,
		ExecutionOrder: view.ExecutionOrder,
		TreeIndex:      view.TreeIndex,
		Position: StepPositionResponse{
			X: view.Position.X,
			Y: view.Position.Y,
		},
		Status:           string(view.Status),
		Attempt:          view.Attempt,
		ResponseSnapshot: view.ResponseSnapshot,
		StartedAt:        view.StartedAt,
		FinishedAt:       view.FinishedAt,
		Error:            optionalNonEmptyString(view.Error),
		CreatedAt:        view.CreatedAt,
		UpdatedAt:        view.UpdatedAt,
	}
}

func NewStepRunListResponseFromViews(views []domainsteprun.StepRunView) []StepRunResponse {
	if len(views) == 0 {
		return nil
	}
	items := make([]StepRunResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewStepRunResponseFromView(view))
	}
	return items
}
