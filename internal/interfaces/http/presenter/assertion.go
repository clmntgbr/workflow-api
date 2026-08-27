package presenter

import (
	"time"

	domainassertion "go-api/internal/domain/assertion"
)

type AssertionResponse struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Source        string    `json:"source"`
	Path          string    `json:"path,omitempty"`
	Operator      string    `json:"operator"`
	ExpectedValue *string   `json:"expectedValue,omitempty"`
	StepID        string    `json:"stepId"`
	WorkflowID    string    `json:"workflowId"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func NewAssertionResponseFromEntity(a domainassertion.Assertion) AssertionResponse {
	return NewAssertionResponseFromView(domainassertion.AssertionView{
		ID:            a.ID,
		Name:          a.Name,
		Description:   a.Description,
		Source:        a.Source,
		Path:          a.Path,
		Operator:      a.Operator,
		ExpectedValue: a.ExpectedValue,
		StepID:        a.StepID,
		WorkflowID:    a.WorkflowID,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	})
}

func NewAssertionResponseFromView(view domainassertion.AssertionView) AssertionResponse {
	var expected *string
	if view.ExpectedValue != "" {
		v := view.ExpectedValue
		expected = &v
	}
	return AssertionResponse{
		ID:            view.ID.String(),
		Name:          view.Name,
		Description:   view.Description,
		Source:        string(view.Source),
		Path:          view.Path,
		Operator:      string(view.Operator),
		ExpectedValue: expected,
		StepID:        view.StepID.String(),
		WorkflowID:    view.WorkflowID.String(),
		CreatedAt:     view.CreatedAt,
		UpdatedAt:     view.UpdatedAt,
	}
}

func NewAssertionListResponseFromViews(views []domainassertion.AssertionView) []AssertionResponse {
	items := make([]AssertionResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewAssertionResponseFromView(view))
	}
	return items
}

type AssertionResultResponse struct {
	AssertionID string `json:"assertionId"`
	Name        string `json:"name"`
	Passed      bool   `json:"passed"`
	ActualValue any    `json:"actualValue"`
	Message     string `json:"message,omitempty"`
}

func NewAssertionResultListResponse(results []domainassertion.Result) []AssertionResultResponse {
	if len(results) == 0 {
		return []AssertionResultResponse{}
	}
	items := make([]AssertionResultResponse, 0, len(results))
	for _, r := range results {
		items = append(items, AssertionResultResponse{
			AssertionID: r.AssertionID,
			Name:        r.Name,
			Passed:      r.Passed,
			ActualValue: r.ActualValue,
			Message:     r.Message,
		})
	}
	return items
}
