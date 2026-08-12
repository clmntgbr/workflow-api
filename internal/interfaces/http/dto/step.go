package dto

type PositionRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CreateStepRequest struct {
	EndpointID string          `json:"endpointId" validate:"required,uuid"`
	Index      string          `json:"index" validate:"required,min=1,max=255"`
	Position   PositionRequest `json:"position"`
}
