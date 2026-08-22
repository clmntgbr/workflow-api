package dto

type StartWorkflowRunRequest struct {
	Context map[string]any `json:"context" validate:"omitempty"`
}

type WorkflowRunAnalyticsQuery struct {
	From string `query:"from" validate:"omitempty"`
	To   string `query:"to" validate:"omitempty"`
}
