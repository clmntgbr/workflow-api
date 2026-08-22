package write

import (
	"time"

	domainproject "go-api/internal/domain/project"

	"github.com/google/uuid"
)

type ProjectModel struct {
	ID        uuid.UUID `gorm:"column:id;primaryKey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ProjectModel) TableName() string {
	return "projects"
}

type UserProjectModel struct {
	UserID         uuid.UUID `gorm:"column:user_id;primaryKey"`
	ProjectID uuid.UUID `gorm:"column:project_id;primaryKey"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (UserProjectModel) TableName() string {
	return "user_projects"
}

func projectModelFromDomain(o *domainproject.Project) *ProjectModel {
	return &ProjectModel{
		ID:        o.ID,
		Name:      o.Name,
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}
}

func projectDomainFromModel(m *ProjectModel, memberIDs []uuid.UUID) *domainproject.Project {
	return &domainproject.Project{
		ID:        m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		MemberIDs: memberIDs,
	}
}
