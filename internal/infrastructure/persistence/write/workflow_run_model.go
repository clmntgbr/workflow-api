package write

import (
	"encoding/json"
	"time"

	domainworkflowrun "go-api/internal/domain/workflowrun"
	"go-api/internal/infrastructure/persistence/dbtype"

	"github.com/google/uuid"
)

type WorkflowRunModel struct {
	ID                uuid.UUID    `gorm:"column:id;primaryKey"`
	WorkflowID        uuid.UUID    `gorm:"column:workflow_id"`
	Status            string       `gorm:"column:status"`
	TriggeredBy       string       `gorm:"column:triggered_by"`
	TriggeredByUserID *uuid.UUID   `gorm:"column:triggered_by_user_id"`
	Context           dbtype.JSONB `gorm:"column:context"`
	StartedAt         *time.Time   `gorm:"column:started_at"`
	FinishedAt        *time.Time   `gorm:"column:finished_at"`
	Error             string       `gorm:"column:error"`
	CreatedAt         time.Time    `gorm:"column:created_at"`
	UpdatedAt         time.Time    `gorm:"column:updated_at"`
}

func (WorkflowRunModel) TableName() string {
	return "workflow_runs"
}

func workflowRunModelFromDomain(r *domainworkflowrun.WorkflowRun) (*WorkflowRunModel, error) {
	ctx := r.Context
	if ctx == nil {
		ctx = map[string]any{}
	}
	contextRaw, err := json.Marshal(ctx)
	if err != nil {
		return nil, err
	}

	return &WorkflowRunModel{
		ID:                r.ID,
		WorkflowID:        r.WorkflowID,
		Status:            string(r.Status),
		TriggeredBy:       string(r.TriggeredBy),
		TriggeredByUserID: r.TriggeredByUserID,
		Context:           dbtype.JSONB(contextRaw),
		StartedAt:         r.StartedAt,
		FinishedAt:        r.FinishedAt,
		Error:             r.Error,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}, nil
}

func workflowRunDomainFromModel(m *WorkflowRunModel) (*domainworkflowrun.WorkflowRun, error) {
	ctx := map[string]any{}
	if len(m.Context) > 0 {
		if err := json.Unmarshal(m.Context, &ctx); err != nil {
			return nil, err
		}
	}

	return &domainworkflowrun.WorkflowRun{
		ID:                m.ID,
		WorkflowID:        m.WorkflowID,
		Status:            domainworkflowrun.Status(m.Status),
		TriggeredBy:       domainworkflowrun.TriggeredBy(m.TriggeredBy),
		TriggeredByUserID: m.TriggeredByUserID,
		Context:           ctx,
		StartedAt:         m.StartedAt,
		FinishedAt:        m.FinishedAt,
		Error:             m.Error,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}, nil
}
