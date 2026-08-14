package presenter

import (
	"time"

	domainvariable "go-api/internal/domain/variable"
)

type VariableResponse struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Key        string    `json:"key"`
	Description *string  `json:"description"`
	Path       string    `json:"path"`
	StepID     string    `json:"stepId"`
	WorkflowID string    `json:"workflowId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

func NewVariableResponseFromEntity(v domainvariable.Variable) VariableResponse {
	return NewVariableResponseFromView(domainvariable.VariableView{
		ID:         v.ID,
		Name:       v.Name,
		Key:        v.Key,
		Description: v.Description,
		Path:       v.Path,
		StepID:     v.StepID,
		WorkflowID: v.WorkflowID,
		CreatedAt:  v.CreatedAt,
		UpdatedAt:  v.UpdatedAt,
	})
}

func NewVariableResponseFromView(view domainvariable.VariableView) VariableResponse {
	return VariableResponse{
		ID:         view.ID.String(),
		Name:       view.Name,
		Key:        view.Key,
		Description: optionalNonEmptyString(view.Description),
		Path:       view.Path,
		StepID:     view.StepID.String(),
		WorkflowID: view.WorkflowID.String(),
		CreatedAt:  view.CreatedAt,
		UpdatedAt:  view.UpdatedAt,
	}
}

func NewVariableListResponseFromViews(views []domainvariable.VariableView) []VariableResponse {
	items := make([]VariableResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewVariableResponseFromView(view))
	}
	return items
}
