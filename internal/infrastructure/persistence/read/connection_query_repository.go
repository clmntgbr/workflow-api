package read

import (
	"context"

	domainconnection "go-api/internal/domain/connection"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type connectionReadRow struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	OrganizationID uuid.UUID
	SourceStepID   uuid.UUID
	TargetStepID   uuid.UUID
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
			ID:             row.ID,
			WorkflowID:     row.WorkflowID,
			OrganizationID: row.OrganizationID,
			SourceStepID:   row.SourceStepID,
			TargetStepID:   row.TargetStepID,
		})
	}
	return views, nil
}
