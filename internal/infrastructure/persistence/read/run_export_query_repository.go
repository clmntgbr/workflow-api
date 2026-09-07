package read

import (
	"context"
	"errors"
	"time"

	domainrunexport "go-api/internal/domain/runexport"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type runExportRow struct {
	ID                uuid.UUID
	WorkflowID        uuid.UUID
	ProjectID         uuid.UUID
	RequestedByUserID uuid.UUID
	DateRangeFrom     time.Time
	DateRangeTo       time.Time
	Status            string
	Error             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (runExportRow) TableName() string { return "run_exports" }

type runExportReadRepository struct {
	db *gorm.DB
}

func NewRunExportReadRepository(db *gorm.DB) domainrunexport.ReadRepository {
	return &runExportReadRepository{db: db}
}

func (r *runExportReadRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*domainrunexport.View, error) {
	var row runExportRow
	err := r.db.WithContext(ctx).
		Table("run_exports").
		Where("id = ?", id).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &domainrunexport.View{
		ID:                row.ID,
		WorkflowID:        row.WorkflowID,
		ProjectID:         row.ProjectID,
		RequestedByUserID: row.RequestedByUserID,
		DateRangeFrom:     row.DateRangeFrom,
		DateRangeTo:       row.DateRangeTo,
		Status:            domainrunexport.Status(row.Status),
		Error:             row.Error,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}, nil
}
