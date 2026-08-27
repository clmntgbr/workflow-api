package dto

type CreateAssertionRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=255"`
	Description   string  `json:"description" validate:"omitempty,max=255"`
	Source        string  `json:"source" validate:"required,oneof=status header body"`
	Path          string  `json:"path" validate:"omitempty,max=255"`
	Operator      string  `json:"operator" validate:"required,oneof=equals not_equals not_null is_null contains greater_than less_than matches_regex is_string is_number is_boolean is_array is_object"`
	ExpectedValue *string `json:"expectedValue" validate:"omitempty"`
}

type UpdateAssertionRequest struct {
	Name          string  `json:"name" validate:"required,min=1,max=255"`
	Description   string  `json:"description" validate:"omitempty,max=255"`
	Source        string  `json:"source" validate:"required,oneof=status header body"`
	Path          string  `json:"path" validate:"omitempty,max=255"`
	Operator      string  `json:"operator" validate:"required,oneof=equals not_equals not_null is_null contains greater_than less_than matches_regex is_string is_number is_boolean is_array is_object"`
	ExpectedValue *string `json:"expectedValue" validate:"omitempty"`
}
