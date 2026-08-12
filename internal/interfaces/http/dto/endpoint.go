package dto

type CreateEndpointRequest struct {
	Name           string            `json:"name" validate:"required,min=1,max=255"`
	Description    string            `json:"description" validate:"omitempty,max=2000"`
	URL            string            `json:"url" validate:"required,url,max=2048"`
	Method         string            `json:"method" validate:"required,http_method"`
	Headers        map[string]string `json:"headers" validate:"omitempty"`
	Query          map[string]string `json:"query" validate:"omitempty"`
	Timeout        *int              `json:"timeout" validate:"omitempty,min=1,max=300000"`
	RetryOnFailure *bool             `json:"retryOnFailure" validate:"omitempty"`
	RetryCount     *int              `json:"retryCount" validate:"omitempty,min=0,max=10"`
	RetryDelay     *int              `json:"retryDelay" validate:"omitempty,min=1,max=60000"`
}
