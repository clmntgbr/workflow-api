package presenter

import (
	domainendpoint "go-api/internal/domain/endpoint"
)

type EndpointListResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	URL         string  `json:"url"`
	Method      string  `json:"method"`
	Status      string  `json:"status"`
}

type EndpointDetailResponse struct {
	EndpointListResponse
	Headers        map[string]string `json:"headers"`
	Query          map[string]string `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
}

func NewEndpointListResponseFromView(view domainendpoint.EndpointView) EndpointListResponse {
	return EndpointListResponse{
		ID:          view.ID.String(),
		Name:        view.Name,
		Description: optionalNonEmptyString(view.Description),
		URL:         view.URL,
		Method:      string(view.Method),
		Status:      string(view.Status),
	}
}

func NewEndpointListResponseFromViews(views []domainendpoint.EndpointView) []EndpointListResponse {
	items := make([]EndpointListResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewEndpointListResponseFromView(view))
	}
	return items
}

func NewEndpointDetailResponseFromView(view domainendpoint.EndpointView) EndpointDetailResponse {
	return endpointDetailResponse(
		NewEndpointListResponseFromView(view),
		view.Headers,
		view.Query,
		view.Body,
		view.Timeout,
		view.RetryOnFailure,
		view.RetryCount,
		view.RetryDelay,
	)
}

func NewEndpointDetailResponseFromEntity(e domainendpoint.Endpoint) EndpointDetailResponse {
	return endpointDetailResponse(
		EndpointListResponse{
			ID:          e.ID.String(),
			Name:        e.Name,
			Description: optionalNonEmptyString(e.Description),
			URL:         e.URL,
			Method:      string(e.Method),
			Status:      string(e.Status),
		},
		e.Headers,
		e.Query,
		e.Body,
		e.Timeout,
		e.RetryOnFailure,
		e.RetryCount,
		e.RetryDelay,
	)
}

func endpointDetailResponse(
	list EndpointListResponse,
	headers, query map[string]string,
	body map[string]any,
	timeout int,
	retryOnFailure bool,
	retryCount, retryDelay int,
) EndpointDetailResponse {
	if headers == nil {
		headers = map[string]string{}
	}
	if query == nil {
		query = map[string]string{}
	}
	if body == nil {
		body = map[string]any{}
	}
	return EndpointDetailResponse{
		EndpointListResponse: list,
		Headers:              headers,
		Query:                query,
		Body:                 body,
		Timeout:              timeout,
		RetryOnFailure:       retryOnFailure,
		RetryCount:           retryCount,
		RetryDelay:           retryDelay,
	}
}
