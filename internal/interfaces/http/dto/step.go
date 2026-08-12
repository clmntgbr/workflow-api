package dto

type PositionRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type CreateStepRequest struct {
	EndpointID string          `json:"endpointId" validate:"required,uuid"`
	Index      string          `json:"index" validate:"required,min=1,max=255"`
	Position   PositionRequest `json:"position"`
}

type ListStepsQuery struct {
	WorkflowID string `query:"workflowId" validate:"required,uuid"`
}
