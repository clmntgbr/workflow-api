package write

import (
	"time"

	domainorganization "go-api/internal/domain/organization"

	"github.com/google/uuid"
)

// OrganizationModel is the persistence mapping for table organizations.
// Schema is owned by SQL migrations — do not encode DDL in tags.
type OrganizationModel struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (OrganizationModel) TableName() string {
	return "organizations"
}

// UserOrganizationModel is the persistence mapping for table user_organizations.
type UserOrganizationModel struct {
	UserID         uuid.UUID `gorm:"column:user_id;primaryKey"`
	OrganizationID uuid.UUID `gorm:"column:organization_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (UserOrganizationModel) TableName() string {
	return "user_organizations"
}

func organizationModelFromDomain(o *domainorganization.Organization) *OrganizationModel {
	return &OrganizationModel{
		ID:        o.ID,
		Name:      o.Name,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func organizationDomainFromModel(m *OrganizationModel, memberIDs []uuid.UUID) *domainorganization.Organization {
	return &domainorganization.Organization{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		MemberIDs: memberIDs,
	}
}
