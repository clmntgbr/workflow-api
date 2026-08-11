package read

import (
	"context"
	"errors"
	"time"

	domainorganization "go-api/internal/domain/organization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type organizationRow struct {
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (organizationRow) TableName() string { return "organizations" }

type userOrganizationRow struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
}

func (userOrganizationRow) TableName() string { return "user_organizations" }

type organizationReadRepository struct {
	db *gorm.DB
}

func NewOrganizationReadRepository(db *gorm.DB) domainorganization.OrganizationReadRepository {
	return &organizationReadRepository{db: db}
}

func (r *organizationReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainorganization.OrganizationView, error) {
	var row organizationRow
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

	return toOrganizationView(row, memberIDs), nil
}

func (r *organizationReadRepository) FindByUserID(
	ctx context.Context,
	userID uuid.UUID,
) ([]domainorganization.OrganizationView, error) {
	var memberships []userOrganizationRow
	if err := r.db.WithContext(ctx).
		Select("user_id", "organization_id").
		Where("user_id = ?", userID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	if len(memberships) == 0 {
		return []domainorganization.OrganizationView{}, nil
	}

	orgIDs := make([]uuid.UUID, 0, len(memberships))
	for _, m := range memberships {
		orgIDs = append(orgIDs, m.OrganizationID)
	}

	var rows []organizationRow
	if err := r.db.WithContext(ctx).
		Select("id", "name", "created_at", "updated_at").
		Where("id IN ?", orgIDs).
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	views := make([]domainorganization.OrganizationView, 0, len(rows))
	for _, row := range rows {
		memberIDs, err := r.loadMemberIDs(ctx, row.ID)
		if err != nil {
			return nil, err
		}
		views = append(views, *toOrganizationView(row, memberIDs))
	}
	return views, nil
}

func (r *organizationReadRepository) loadMemberIDs(ctx context.Context, organizationID uuid.UUID) ([]uuid.UUID, error) {
	var memberships []userOrganizationRow
	if err := r.db.WithContext(ctx).
		Select("user_id", "organization_id").
		Where("organization_id = ?", organizationID).
		Find(&memberships).Error; err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(memberships))
	for _, m := range memberships {
		ids = append(ids, m.UserID)
	}
	return ids, nil
}

func toOrganizationView(row organizationRow, memberIDs []uuid.UUID) *domainorganization.OrganizationView {
	return &domainorganization.OrganizationView{
		ID:        row.ID,
		Name:      row.Name,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		MemberIDs: memberIDs,
	}
}
