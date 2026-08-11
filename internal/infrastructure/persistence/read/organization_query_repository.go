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
	IsActive  bool
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
		Select("id", "name", "is_active", "created_at", "updated_at").
		First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	var memberships []userOrganizationRow
	if err := r.db.WithContext(ctx).
		Select("user_id", "organization_id").
		Where("organization_id = ?", id).
		Find(&memberships).Error; err != nil {
		return nil, err
	}

	memberIDs := make([]uuid.UUID, 0, len(memberships))
	for _, m := range memberships {
		memberIDs = append(memberIDs, m.UserID)
	}

	return &domainorganization.OrganizationView{
		ID:        row.ID,
		Name:      row.Name,
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
		MemberIDs: memberIDs,
	}, nil
}
