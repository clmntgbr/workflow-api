package dto

type CreateVariableRequest struct {
	StepID      string `json:"stepId" validate:"required,uuid"`
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Key         string `json:"key" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Path        string `json:"path" validate:"required,min=1,max=255"`
}

type UpdateVariableRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Key         string `json:"key" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Path        string `json:"path" validate:"required,min=1,max=255"`
}
