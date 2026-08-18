package dto

type CreateEndpointRequest struct {
	Name           string            `json:"name" validate:"required,min=1,max=255"`
	Description    string            `json:"description" validate:"omitempty,max=2000"`
	URL            string            `json:"url" validate:"required,url,max=2048"`
	Method         string            `json:"method" validate:"required,http_method"`
	Headers        map[string]string `json:"headers" validate:"omitempty"`
	Query          map[string]string `json:"query" validate:"omitempty"`
	Body           map[string]any    `json:"body" validate:"omitempty"`
	Timeout        *int              `json:"timeout" validate:"required,min=30000,max=300000"`
	RetryOnFailure *bool             `json:"retryOnFailure" validate:"required"`
	RetryCount     *int              `json:"retryCount" validate:"required,min=0,max=10"`
	RetryDelay     *int              `json:"retryDelay" validate:"required,min=10000,max=60000"`
}

type UpdateEndpointRequest struct {
	Name           string            `json:"name" validate:"required,min=1,max=255"`
	Description    string            `json:"description" validate:"omitempty,max=2000"`
	URL            string            `json:"url" validate:"required,url,max=2048"`
	Method         string            `json:"method" validate:"required,http_method"`
	Headers        map[string]string `json:"headers" validate:"omitempty"`
	Query          map[string]string `json:"query" validate:"omitempty"`
	Body           map[string]any    `json:"body" validate:"omitempty"`
	Timeout        *int              `json:"timeout" validate:"required,min=30000,max=300000"`
	RetryOnFailure *bool             `json:"retryOnFailure" validate:"required"`
	RetryCount     *int              `json:"retryCount" validate:"required,min=0,max=10"`
	RetryDelay     *int              `json:"retryDelay" validate:"required,min=10000,max=60000"`
	Status         string            `json:"status" validate:"required,endpoint_status"`
}
