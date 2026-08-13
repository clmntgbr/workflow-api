package dto

type StartWorkflowRunRequest struct {
	Context map[string]any `json:"context" validate:"omitempty"`
}
