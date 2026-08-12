package presenter

import (
	"time"

	domainstep "go-api/internal/domain/step"
)

type StepPositionResponse struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type StepDetailResponse struct {
	ID             string               `json:"id"`
	WorkflowID     string               `json:"workflowId"`
	EndpointID     string               `json:"endpointId"`
	Name           string               `json:"name"`
	Description    *string              `json:"description"`
	URL            string               `json:"url"`
	Method         string               `json:"method"`
	Headers        map[string]string    `json:"headers"`
	Query          map[string]string    `json:"query"`
	Body           map[string]any       `json:"body"`
	Timeout        int                  `json:"timeout"`
	RetryOnFailure bool                 `json:"retryOnFailure"`
	RetryCount     int                  `json:"retryCount"`
	RetryDelay     int                  `json:"retryDelay"`
	Index          string               `json:"index"`
	ExecutionOrder int                  `json:"executionOrder"`
	TreeIndex      int                  `json:"treeIndex"`
	Position       StepPositionResponse `json:"position"`
	Status         string               `json:"status"`
	OrganizationID string               `json:"organizationId"`
	CreatedAt      time.Time            `json:"createdAt"`
	UpdatedAt      time.Time            `json:"updatedAt"`
}

func NewStepDetailResponseFromView(view domainstep.StepView) StepDetailResponse {
	return stepDetailResponse(
		view.ID.String(),
		view.WorkflowID.String(),
		view.EndpointID.String(),
		view.Name,
		view.Description,
		view.URL,
		view.Method,
		view.Headers,
		view.Query,
		view.Body,
		view.Timeout,
		view.RetryOnFailure,
		view.RetryCount,
		view.RetryDelay,
		view.Index,
		view.ExecutionOrder,
		view.TreeIndex,
		view.Position,
		string(view.Status),
		view.OrganizationID.String(),
		view.CreatedAt,
		view.UpdatedAt,
	)
}

func NewStepDetailResponseFromEntity(s domainstep.Step) StepDetailResponse {
	return stepDetailResponse(
		s.ID.String(),
		s.WorkflowID.String(),
		s.EndpointID.String(),
		s.Name,
		s.Description,
		s.URL,
		s.Method,
		s.Headers,
		s.Query,
		s.Body,
		s.Timeout,
		s.RetryOnFailure,
		s.RetryCount,
		s.RetryDelay,
		s.Index,
		s.ExecutionOrder,
		s.TreeIndex,
		s.Position,
		string(s.Status),
		s.OrganizationID.String(),
		s.CreatedAt,
		s.UpdatedAt,
	)
}

func NewStepListResponseFromViews(views []domainstep.StepView) []StepDetailResponse {
	items := make([]StepDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewStepDetailResponseFromView(view))
	}
	return items
}

func stepDetailResponse(
	id, workflowID, endpointID, name, description, url, method string,
	headers, query map[string]string,
	body map[string]any,
	timeout int,
	retryOnFailure bool,
	retryCount, retryDelay int,
	index string,
	executionOrder, treeIndex int,
	position domainstep.Position,
	status, organizationID string,
	createdAt, updatedAt time.Time,
) StepDetailResponse {
	if headers == nil {
		headers = map[string]string{}
	}
	if query == nil {
		query = map[string]string{}
	}
	if body == nil {
		body = map[string]any{}
	}
	return StepDetailResponse{
		ID:             id,
		WorkflowID:     workflowID,
		EndpointID:     endpointID,
		Name:           name,
		Description:    optionalNonEmptyString(description),
		URL:            url,
		Method:         method,
		Headers:        headers,
		Query:          query,
		Body:           body,
		Timeout:        timeout,
		RetryOnFailure: retryOnFailure,
		RetryCount:     retryCount,
		RetryDelay:     retryDelay,
		Index:          index,
		ExecutionOrder: executionOrder,
		TreeIndex:      treeIndex,
		Position: StepPositionResponse{
			X: position.X,
			Y: position.Y,
		},
		Status:         status,
		OrganizationID: organizationID,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}
