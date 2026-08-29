package presenter

import (
	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type StepPositionResponse struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type StepListResponse struct {
	ID                   string               `json:"id"`
	Type                 string               `json:"type"`
	EndpointID           *string              `json:"endpointId,omitempty"`
	DelayDurationSeconds *int                 `json:"delayDurationSeconds,omitempty"`
	Expression           *string              `json:"expression,omitempty"`
	Name                 string               `json:"name"`
	URL                  string               `json:"url"`
	Method               string               `json:"method"`
	Position             StepPositionResponse `json:"position"`
	LastRunStatus        *string              `json:"lastRunStatus"`
}

type StepDetailResponse struct {
	StepListResponse
	Description    *string           `json:"description"`
	Headers        map[string]string `json:"headers"`
	Query          httpquery.Params  `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
	Index          *string           `json:"index"`
	ExecutionOrder int               `json:"executionOrder"`
	TreeIndex      int               `json:"treeIndex"`
	Status         string            `json:"status"`
}

func NewStepListResponseFromView(view domainstep.StepView) StepListResponse {
	return StepListResponse{
		ID:                   view.ID.String(),
		Type:                 string(view.Type),
		EndpointID:           optionalUUIDString(view.EndpointID),
		DelayDurationSeconds: optionalPositiveInt(view.DelayDurationSeconds),
		Expression:           view.Expression,
		Name:                 view.Name,
		URL:                  view.URL,
		Method:               view.Method,
		Position: StepPositionResponse{
			X: view.Position.X,
			Y: view.Position.Y,
		},
	}
}

func optionalPositiveInt(value int) *int {
	if value <= 0 {
		return nil
	}
	return &value
}

func NewStepListResponseFromViews(views []domainstep.StepView) []StepListResponse {
	return NewStepListResponseFromViewsWithLastRunStatus(views, nil)
}

func NewStepListResponseFromViewsWithLastRunStatus(
	views []domainstep.StepView,
	lastRunStatusByStepID map[uuid.UUID]string,
) []StepListResponse {
	items := make([]StepListResponse, 0, len(views))
	for _, view := range views {
		item := NewStepListResponseFromView(view)
		if lastRunStatusByStepID != nil {
			if status, ok := lastRunStatusByStepID[view.ID]; ok && status != "" {
				value := status
				item.LastRunStatus = &value
			}
		}
		items = append(items, item)
	}
	return items
}

func NewStepDetailResponseFromView(view domainstep.StepView) StepDetailResponse {
	return stepDetailResponse(
		NewStepListResponseFromView(view),
		view.Description,
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
		string(view.Status),
	)
}

func NewStepDetailResponseFromEntity(s domainstep.Step) StepDetailResponse {
	return stepDetailResponse(
		StepListResponse{
			ID:                   s.ID.String(),
			Type:                 string(s.Type),
			EndpointID:           optionalUUIDString(s.EndpointID),
			DelayDurationSeconds: optionalPositiveInt(s.DelayDurationSeconds),
			Expression:           s.Expression,
			Name:                 s.Name,
			URL:                  s.URL,
			Method:               s.Method,
			Position: StepPositionResponse{
				X: s.Position.X,
				Y: s.Position.Y,
			},
		},
		s.Description,
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
		string(s.Status),
	)
}

func stepDetailResponse(
	list StepListResponse,
	description string,
	headers map[string]string,
	query httpquery.Params,
	body map[string]any,
	timeout int,
	retryOnFailure bool,
	retryCount, retryDelay int,
	index string,
	executionOrder, treeIndex int,
	status string,
) StepDetailResponse {
	if headers == nil {
		headers = map[string]string{}
	}
	if query == nil {
		query = httpquery.Empty()
	}
	if body == nil {
		body = map[string]any{}
	}
	return StepDetailResponse{
		StepListResponse: list,
		Description:      optionalNonEmptyString(description),
		Headers:          headers,
		Query:            query,
		Body:             body,
		Timeout:          timeout,
		RetryOnFailure:   retryOnFailure,
		RetryCount:       retryCount,
		RetryDelay:       retryDelay,
		Index:            optionalNonEmptyString(index),
		ExecutionOrder:   executionOrder,
		TreeIndex:        treeIndex,
		Status:           status,
	}
}
