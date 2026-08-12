package presenter

import (
	"time"

	domainendpoint "go-api/internal/domain/endpoint"
)

type EndpointDetailResponse struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	Description    *string           `json:"description"`
	URL            string            `json:"url"`
	Method         string            `json:"method"`
	Headers        map[string]string `json:"headers"`
	Query          map[string]string `json:"query"`
	Body           map[string]any    `json:"body"`
	Timeout        int               `json:"timeout"`
	RetryOnFailure bool              `json:"retryOnFailure"`
	RetryCount     int               `json:"retryCount"`
	RetryDelay     int               `json:"retryDelay"`
	Status         string            `json:"status"`
	OrganizationID string            `json:"organizationId"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

func NewEndpointDetailResponseFromView(view domainendpoint.EndpointView) EndpointDetailResponse {
	return endpointDetailResponse(
		view.ID.String(),
		view.Name,
		view.Description,
		view.URL,
		string(view.Method),
		view.Headers,
		view.Query,
		view.Body,
		view.Timeout,
		view.RetryOnFailure,
		view.RetryCount,
		view.RetryDelay,
		string(view.Status),
		view.OrganizationID.String(),
		view.CreatedAt,
		view.UpdatedAt,
	)
}

func NewEndpointDetailResponseFromEntity(e domainendpoint.Endpoint) EndpointDetailResponse {
	return endpointDetailResponse(
		e.ID.String(),
		e.Name,
		e.Description,
		e.URL,
		string(e.Method),
		e.Headers,
		e.Query,
		e.Body,
		e.Timeout,
		e.RetryOnFailure,
		e.RetryCount,
		e.RetryDelay,
		string(e.Status),
		e.OrganizationID.String(),
		e.CreatedAt,
		e.UpdatedAt,
	)
}

func NewEndpointListResponseFromViews(views []domainendpoint.EndpointView) []EndpointDetailResponse {
	items := make([]EndpointDetailResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewEndpointDetailResponseFromView(view))
	}
	return items
}

func endpointDetailResponse(
	id, name, description, url, method string,
	headers, query map[string]string,
	body map[string]any,
	timeout int,
	retryOnFailure bool,
	retryCount, retryDelay int,
	status, organizationID string,
	createdAt, updatedAt time.Time,
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
		ID:             id,
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
		Status:         status,
		OrganizationID: organizationID,
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
	}
}
