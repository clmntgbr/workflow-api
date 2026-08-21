package dto

type CreateVariableRequest struct {
	StepID      *string `json:"stepId" validate:"omitempty,uuid"`
	Kind        string  `json:"kind" validate:"omitempty,oneof=extracted static"`
	Name        string  `json:"name" validate:"required,min=1,max=255"`
	Key         string  `json:"key" validate:"required,min=1,max=255"`
	Description string  `json:"description" validate:"omitempty,max=255"`
	Path        string  `json:"path" validate:"omitempty,max=255"`
	Value       any     `json:"value"`
}

type UpdateVariableRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Key         string `json:"key" validate:"required,min=1,max=255"`
	Description string `json:"description" validate:"omitempty,max=255"`
	Path        string `json:"path" validate:"omitempty,max=255"`
	Value       any    `json:"value"`
}
