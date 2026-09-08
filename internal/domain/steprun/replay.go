package steprun

import (
	"time"

	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/httpquery"

	"github.com/google/uuid"
)

func (v StepRunView) ReplayOnto(workflowRunID uuid.UUID) *StepRun {
	now := time.Now().UTC()
	cloned := &StepRun{
		ID:                   uuid.New(),
		WorkflowRunID:        workflowRunID,
		StepID:               v.StepID,
		WorkflowID:           v.WorkflowID,
		ProjectID:            v.ProjectID,
		StepType:             v.StepType,
		DelayDurationSeconds: v.DelayDurationSeconds,
		Name:                 v.Name,
		Description:          v.Description,
		URL:                  v.URL,
		Method:               v.Method,
		Headers:              normalizeStringMap(v.Headers),
		Query:                httpquery.Clone(v.Query),
		Body:                 normalizeAnyMap(v.Body),
		Timeout:              v.Timeout,
		RetryOnFailure:       v.RetryOnFailure,
		RetryCount:           v.RetryCount,
		RetryDelay:           v.RetryDelay,
		Index:                v.Index,
		ExecutionOrder:       v.ExecutionOrder,
		TreeIndex:            v.TreeIndex,
		Position:             v.Position,
		Status:               v.Status,
		Attempt:              v.Attempt,
		VariableExtracts:     append([]VariableExtract(nil), v.VariableExtracts...),
		Assertions:           append([]domainassertion.Snapshot(nil), v.Assertions...),
		ExtractedVariables:   normalizeAnyMap(v.ExtractedVariables),
		AssertionsResult:     append([]domainassertion.Result(nil), v.AssertionsResult...),
		StartedAt:            cloneTime(v.StartedAt),
		FinishedAt:           cloneTime(v.FinishedAt),
		ResumeAt:             cloneTime(v.ResumeAt),
		MatchedBranch:        cloneBool(v.MatchedBranch),
		Error:                v.Error,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if v.ResponseSnapshot != nil {
		normalized := v.ResponseSnapshot.Normalized()
		cloned.ResponseSnapshot = &normalized
	}
	return cloned
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}

func cloneBool(value *bool) *bool {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
