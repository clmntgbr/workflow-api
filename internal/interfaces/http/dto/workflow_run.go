package dto

type StartWorkflowRunRequest struct {
	Context    map[string]any `json:"context" validate:"omitempty"`
	FromStepID *string        `json:"fromStepId" validate:"omitempty,uuid"`
}

type WorkflowRunAnalyticsQuery struct {
	From string `query:"from" validate:"omitempty"`
	To   string `query:"to" validate:"omitempty"`
}
