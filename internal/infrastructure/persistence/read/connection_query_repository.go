package read

import (
	"context"

	domainconnection "go-api/internal/domain/connection"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type connectionReadRow struct {
	ID           uuid.UUID
	WorkflowID   uuid.UUID
	ProjectID    uuid.UUID
	SourceStepID uuid.UUID
	TargetStepID uuid.UUID
	Branch       *string
}

func (connectionReadRow) TableName() string { return "connections" }

type connectionReadRepository struct {
	db *gorm.DB
}

func NewConnectionReadRepository(db *gorm.DB) domainconnection.ConnectionReadRepository {
	return &connectionReadRepository{db: db}
}

func (r *connectionReadRepository) FindByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]domainconnection.ConnectionView, error) {
	var rows []connectionReadRow
	err := r.db.WithContext(ctx).
		Where("workflow_id = ?", workflowID).
		Order("id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	views := make([]domainconnection.ConnectionView, 0, len(rows))
	for _, row := range rows {
		views = append(views, domainconnection.ConnectionView{
			ID:           row.ID,
			WorkflowID:   row.WorkflowID,
			ProjectID:    row.ProjectID,
			SourceStepID: row.SourceStepID,
			TargetStepID: row.TargetStepID,
			Branch:       branchFromReadRow(row.Branch),
		})
	}
	return views, nil
}

func branchFromReadRow(branch *string) *domainconnection.ConditionBranch {
	if branch == nil {
		return nil
	}
	value := domainconnection.ConditionBranch(*branch)
	return &value
}
