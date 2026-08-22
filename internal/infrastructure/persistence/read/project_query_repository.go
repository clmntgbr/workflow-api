package read

import (
	"context"
	"errors"
	"time"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRow struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (projectRow) TableName() string { return "projects" }

type userProjectRow struct {
	UserID         uuid.UUID
	ProjectID uuid.UUID
}

func (userProjectRow) TableName() string { return "user_projects" }

type projectReadRepository struct {
	db *gorm.DB
}

func NewProjectReadRepository(db *gorm.DB) domainproject.ProjectReadRepository {
	return &projectReadRepository{db: db}
}

func (r *projectReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainproject.ProjectView, error) {
	var row projectRow
	err := r.db.WithContext(ctx).
		Select("id", "name", "created_at", "updated_at").
		First(&row, "id = ?", id).Error
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

	return toProjectView(row, memberIDs), nil
}

func (r *projectReadRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]domainproject.ProjectView, error) {
	var memberships []userProjectRow
	if err := r.db.WithContext(ctx).
		Select("user_id", "project_id").
		Where("user_id = ?", userID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return []domainproject.ProjectView{}, nil
	}

	orgIDs := make([]uuid.UUID, 0, len(memberships))
	for _, m := range memberships {
		orgIDs = append(orgIDs, m.ProjectID)
	}

	var rows []projectRow
	if err := r.db.WithContext(ctx).
		Select("id", "name", "created_at", "updated_at").
		Where("id IN ?", orgIDs).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	views := make([]domainproject.ProjectView, 0, len(rows))
	for _, row := range rows {
		memberIDs, err := r.loadMemberIDs(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		views = append(views, *toProjectView(row, memberIDs))
	}
	return views, nil
}

func (r *projectReadRepository) loadMemberIDs(ctx context.Context, projectID uuid.UUID) ([]uuid.UUID, error) {
	var memberships []userProjectRow
	if err := r.db.WithContext(ctx).
		Select("user_id", "project_id").
		Where("project_id = ?", projectID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(memberships))
	for _, m := range memberships {
		ids = append(ids, m.UserID)
	}
	return ids, nil
}

func toProjectView(row projectRow, memberIDs []uuid.UUID) *domainproject.ProjectView {
	return &domainproject.ProjectView{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		MemberIDs: memberIDs,
	}
}
