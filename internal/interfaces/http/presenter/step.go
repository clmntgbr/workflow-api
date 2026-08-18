package presenter

import domainstep "go-api/internal/domain/step"

type StepPositionResponse struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type StepListResponse struct {
	ID         string               `json:"id"`
	EndpointID string               `json:"endpointId"`
	Name       string               `json:"name"`
	URL        string               `json:"url"`
	Method     string               `json:"method"`
	Position   StepPositionResponse `json:"position"`
}

type StepDetailResponse struct {
	StepListResponse
	Description    *string           `json:"description"`
	Headers        map[string]string `json:"headers"`
	Query          map[string]string `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
	Index          string            `json:"index"`
	ExecutionOrder int               `json:"executionOrder"`
	TreeIndex      int               `json:"treeIndex"`
	Status         string            `json:"status"`
}

func NewStepListResponseFromView(view domainstep.StepView) StepListResponse {
	return StepListResponse{
		ID:         view.ID.String(),
		EndpointID: view.EndpointID.String(),
		Name:       view.Name,
		URL:        view.URL,
		Method:     view.Method,
		Position: StepPositionResponse{
			X: view.Position.X,
			Y: view.Position.Y,
		},
	}
}

func NewStepListResponseFromViews(views []domainstep.StepView) []StepListResponse {
	items := make([]StepListResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewStepListResponseFromView(view))
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
			ID:         s.ID.String(),
			EndpointID: s.EndpointID.String(),
			Name:       s.Name,
			URL:        s.URL,
			Method:     s.Method,
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
	headers, query map[string]string,
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
		query = map[string]string{}
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
		Index:            index,
		ExecutionOrder:   executionOrder,
		TreeIndex:        treeIndex,
		Status:           status,
	}
}
