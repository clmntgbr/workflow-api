package dto

type CreateConnectionRequest struct {
	SourceStepID string `json:"sourceStepId" validate:"required,uuid"`
	TargetStepID string `json:"targetStepId" validate:"required,uuid"`
}
