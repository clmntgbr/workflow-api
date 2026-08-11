package write

import (
	"context"
	"errors"
	"time"

	domainorganization "go-api/internal/domain/organization"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type organizationWriteRepository struct {
	db *gorm.DB
}

func NewOrganizationWriteRepository(db *gorm.DB) domainorganization.OrganizationWriteRepository {
	return &organizationWriteRepository{db: db}
}

func (r *organizationWriteRepository) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(ContextWithTx(ctx, tx))
	})
}

func (r *organizationWriteRepository) Save(ctx context.Context, org *domainorganization.Organization) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Create(organizationModelFromDomain(org)).Error; err != nil {
		return err
	}
	return r.ReplaceMembers(ctx, org.ID, org.MemberIDs)
}

func (r *organizationWriteRepository) Update(ctx context.Context, org *domainorganization.Organization) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Save(organizationModelFromDomain(org)).Error; err != nil {
		return err
	}
	return r.ReplaceMembers(ctx, org.ID, org.MemberIDs)
}

func (r *organizationWriteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("organization_id = ?", id).Delete(&UserOrganizationModel{}).Error; err != nil {
		return err
	}
	return db.Delete(&OrganizationModel{}, id).Error
}

func (r *organizationWriteRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainorganization.Organization, error) {
	db := DBWithContext(ctx, r.db)
	var model OrganizationModel
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
	return organizationDomainFromModel(&model, memberIDs), nil
}

func (r *organizationWriteRepository) ReplaceMembers(ctx context.Context, organizationID uuid.UUID, memberIDs []uuid.UUID) error {
	db := DBWithContext(ctx, r.db)
	if err := db.Where("organization_id = ?", organizationID).Delete(&UserOrganizationModel{}).Error; err != nil {
		return err
	}
	if len(memberIDs) == 0 {
		return nil
	}

	now := time.Now().UTC()
	rows := make([]UserOrganizationModel, 0, len(memberIDs))
	for _, userID := range memberIDs {
		rows = append(rows, UserOrganizationModel{
			UserID:         userID,
			OrganizationID: organizationID,
			CreatedAt:      now,
		})
	}
	return db.Create(&rows).Error
}

func (r *organizationWriteRepository) loadMemberIDs(ctx context.Context, organizationID uuid.UUID) ([]uuid.UUID, error) {
	var rows []UserOrganizationModel
	err := DBWithContext(ctx, r.db).
		Where("organization_id = ?", organizationID).
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
