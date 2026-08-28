package presenter

import (
	"time"

	domainassertion "go-api/internal/domain/assertion"

	"github.com/google/uuid"
)

type AssertionResponse struct {
	ID            string     `json:"id"`
	Description   *string    `json:"description"`
	Source        string     `json:"source"`
	Path          *string    `json:"path"`
	Operator      string     `json:"operator"`
	ExpectedValue *string    `json:"expectedValue"`
	StepID        string     `json:"stepId"`
	WorkflowID    string     `json:"workflowId"`
	CreatedAt     *time.Time `json:"createdAt"`
	UpdatedAt     *time.Time `json:"updatedAt"`
}

func NewAssertionResponseFromEntity(a domainassertion.Assertion) AssertionResponse {
	return NewAssertionResponseFromView(domainassertion.AssertionView{
		ID:            a.ID,
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
	return AssertionResponse{
		ID:            view.ID.String(),
		Description:   optionalNonEmptyString(view.Description),
		Source:        string(view.Source),
		Path:          optionalNonEmptyString(view.Path),
		Operator:      string(view.Operator),
		ExpectedValue: optionalNonEmptyString(view.ExpectedValue),
		StepID:        view.StepID.String(),
		WorkflowID:    view.WorkflowID.String(),
		CreatedAt:     optionalTime(view.CreatedAt),
		UpdatedAt:     optionalTime(view.UpdatedAt),
	}
}

func NewAssertionResponseFromSnapshot(
	snapshot domainassertion.Snapshot,
	stepID uuid.UUID,
	workflowID uuid.UUID,
) AssertionResponse {
	return AssertionResponse{
		ID:            snapshot.AssertionID,
		Description:   optionalNonEmptyString(snapshot.Description),
		Source:        string(snapshot.Source),
		Path:          optionalNonEmptyString(snapshot.Path),
		Operator:      string(snapshot.Operator),
		ExpectedValue: optionalNonEmptyString(snapshot.ExpectedValue),
		StepID:        stepID.String(),
		WorkflowID:    workflowID.String(),
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
	Assertion   AssertionResponse `json:"assertion"`
	Passed      bool              `json:"passed"`
	ActualValue any               `json:"actualValue"`
	Message     *string           `json:"message"`
}

func NewAssertionResultListResponse(
	results []domainassertion.Result,
	snapshots []domainassertion.Snapshot,
	stepID uuid.UUID,
	workflowID uuid.UUID,
) []AssertionResultResponse {
	if len(results) == 0 {
		return []AssertionResultResponse{}
	}

	snapshotsByID := make(map[string]domainassertion.Snapshot, len(snapshots))
	for _, snapshot := range snapshots {
		snapshotsByID[snapshot.AssertionID] = snapshot
	}

	items := make([]AssertionResultResponse, 0, len(results))
	for _, result := range results {
		snapshot, ok := snapshotsByID[result.AssertionID]
		var assertion AssertionResponse
		if ok {
			assertion = NewAssertionResponseFromSnapshot(snapshot, stepID, workflowID)
		} else {
			assertion = AssertionResponse{
				ID:         result.AssertionID,
				StepID:     stepID.String(),
				WorkflowID: workflowID.String(),
			}
		}
		items = append(items, AssertionResultResponse{
			Assertion:   assertion,
			Passed:      result.Passed,
			ActualValue: result.ActualValue,
			Message:     optionalNonEmptyString(result.Message),
		})
	}
	return items
}
