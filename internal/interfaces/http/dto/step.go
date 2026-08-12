package dto

type PositionRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type CreateStepRequest struct {
	EndpointID string          `json:"endpointId" validate:"required,uuid"`
	Position   PositionRequest `json:"position"`
}

type UpdateStepPositionRequest struct {
	Position PositionRequest `json:"position"`
}
