package dto

import "go-api/internal/domain/httpquery"

type PositionRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CreateStepRequest struct {
	Type                 string          `json:"type" validate:"omitempty,oneof=http delay"`
	EndpointID           string          `json:"endpointId" validate:"omitempty,uuid"`
	Name                 string          `json:"name" validate:"omitempty,min=1,max=255"`
	DelayDurationSeconds *int            `json:"delayDurationSeconds" validate:"omitempty,min=1"`
	Position             PositionRequest `json:"position"`
}

type UpdateDelayStepRequest struct {
	Name                 string `json:"name" validate:"required,min=1,max=255"`
	Description          string `json:"description" validate:"omitempty,max=2000"`
	DelayDurationSeconds int    `json:"delayDurationSeconds" validate:"required,min=1"`
}

type UpdateStepPositionRequest struct {
	Position PositionRequest `json:"position"`
}

type UpdateStepRequest struct {
	Name           string            `json:"name" validate:"required,min=1,max=255"`
	Description    string            `json:"description" validate:"omitempty,max=2000"`
	URL            string            `json:"url" validate:"required,max=2048"`
	Method         string            `json:"method" validate:"required,http_method"`
	Headers        map[string]string `json:"headers" validate:"omitempty"`
	Query          httpquery.Params  `json:"query" validate:"omitempty"`
	Body           map[string]any    `json:"body" validate:"omitempty"`
	Timeout        *int              `json:"timeout" validate:"required,min=30000,max=300000"`
	RetryOnFailure *bool             `json:"retryOnFailure" validate:"required"`
	RetryCount     *int              `json:"retryCount" validate:"required,min=0,max=10"`
	RetryDelay     *int              `json:"retryDelay" validate:"required,min=10000,max=60000"`
}
