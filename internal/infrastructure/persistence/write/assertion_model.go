package write

import (
	"time"

	domainassertion "go-api/internal/domain/assertion"

	"github.com/google/uuid"
)

type AssertionModel struct {
	ID            uuid.UUID `gorm:"column:id;primaryKey"`
	Name          string    `gorm:"column:name"`
	Description   string    `gorm:"column:description"`
	Source        string    `gorm:"column:source"`
	Path          *string   `gorm:"column:path"`
	Operator      string    `gorm:"column:operator"`
	ExpectedValue *string   `gorm:"column:expected_value"`
	StepID        uuid.UUID `gorm:"column:step_id"`
	WorkflowID    uuid.UUID `gorm:"column:workflow_id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (AssertionModel) TableName() string { return "assertions" }

func assertionModelFromDomain(a *domainassertion.Assertion) *AssertionModel {
	var path *string
	if a.Path != "" {
		p := a.Path
		path = &p
	}
	var expected *string
	if a.ExpectedValue != "" {
		v := a.ExpectedValue
		expected = &v
	}
	return &AssertionModel{
		ID:            a.ID,
		Name:          a.Name,
		Description:   a.Description,
		Source:        string(a.Source),
		Path:          path,
		Operator:      string(a.Operator),
		ExpectedValue: expected,
		StepID:        a.StepID,
		WorkflowID:    a.WorkflowID,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

func assertionDomainFromModel(m *AssertionModel) *domainassertion.Assertion {
	path := ""
	if m.Path != nil {
		path = *m.Path
	}
	expected := ""
	if m.ExpectedValue != nil {
		expected = *m.ExpectedValue
	}
	return &domainassertion.Assertion{
		ID:            m.ID,
		Name:          m.Name,
		Description:   m.Description,
		Source:        domainassertion.AssertionSource(m.Source),
		Path:          path,
		Operator:      domainassertion.AssertionOperator(m.Operator),
		ExpectedValue: expected,
		StepID:        m.StepID,
		WorkflowID:    m.WorkflowID,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
