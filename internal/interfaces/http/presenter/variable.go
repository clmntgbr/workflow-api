package presenter

import domainvariable "go-api/internal/domain/variable"

type VariableResponse struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Key         string  `json:"key"`
	Description *string `json:"description"`
	Kind        string  `json:"kind"`
	Path        *string `json:"path"`
	Value       any     `json:"value"`
	StepID      *string `json:"stepId"`
}

func NewVariableResponseFromEntity(v domainvariable.Variable) VariableResponse {
	return NewVariableResponseFromView(domainvariable.VariableView{
		ID:          v.ID,
		Name:        v.Name,
		Key:         v.Key,
		Description: v.Description,
		Kind:        v.Kind,
		Path:        v.Path,
		Value:       v.Value,
		StepID:      v.StepID,
		WorkflowID:  v.WorkflowID,
	})
}

func NewVariableResponseFromView(view domainvariable.VariableView) VariableResponse {
	kind := string(view.Kind)
	if kind == "" {
		kind = string(domainvariable.KindExtracted)
	}

	var path *string
	if view.Kind != domainvariable.KindStatic {
		path = optionalNonEmptyString(view.Path)
	}

	return VariableResponse{
		ID:          view.ID.String(),
		Name:        view.Name,
		Key:         view.Key,
		Description: optionalNonEmptyString(view.Description),
		Kind:        kind,
		Path:        path,
		Value:       view.Value,
		StepID:      optionalUUIDString(view.StepID),
	}
}

func NewVariableListResponseFromViews(views []domainvariable.VariableView) []VariableResponse {
	items := make([]VariableResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewVariableResponseFromView(view))
	}
	return items
}
