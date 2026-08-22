package write

import (
	"context"
	"errors"
	"time"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectWriteRepository struct {
	db *gorm.DB
}

func NewProjectWriteRepository(db *gorm.DB) domainproject.ProjectWriteRepository {
	return &projectWriteRepository{db: db}
}

func (r *projectWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *projectWriteRepository) Save(ctx context.Context, org *domainproject.Project) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Create(projectModelFromDomain(org)).Error; err != nil {
		return err
	}
	return r.ReplaceMembers(ctx, org.ID, org.MemberIDs)
}

func (r *projectWriteRepository) Update(ctx context.Context, org *domainproject.Project) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Save(projectModelFromDomain(org)).Error; err != nil {
		return err
	}
	return r.ReplaceMembers(ctx, org.ID, org.MemberIDs)
}

func (r *projectWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("project_id = ?", id).Delete(&UserProjectModel{}).Error; err != nil {
		return err
	}
	return db.Delete(&ProjectModel{}, id).Error
}

func (r *projectWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainproject.Project, error) {
	db := DBWithContext(ctx, r.db)
	var model ProjectModel
	err := db.First(&model, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	memberIDs, err := r.loadMemberIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	return projectDomainFromModel(&model, memberIDs), nil
}

func (r *projectWriteRepository) ReplaceMembers(ctx context.Context, projectID uuid.UUID, memberIDs []uuid.UUID) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("project_id = ?", projectID).Delete(&UserProjectModel{}).Error; err != nil {
		return err
	}
	if len(memberIDs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	rows := make([]UserProjectModel, 0, len(memberIDs))
	for _, userID := range memberIDs {
		rows = append(rows, UserProjectModel{
			UserID:         userID,
			ProjectID: projectID,
			CreatedAt:      now,
		})
	}
	return db.Create(&rows).Error
}

func (r *projectWriteRepository) loadMemberIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	var rows []UserProjectModel
	err := DBWithContext(ctx, r.db).
		Where("project_id = ?", projectID).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserID)
	}
	return ids, nil
}
