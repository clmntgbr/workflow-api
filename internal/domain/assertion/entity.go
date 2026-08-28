package assertion

import (
	"strings"
	"time"

	"go-api/internal/domain/event"

	"github.com/google/uuid"
)

type Assertion struct {
	ID            uuid.UUID
	Description   string
	Source        AssertionSource
	Path          string
	Operator      AssertionOperator
	ExpectedValue string
	StepID        uuid.UUID
	WorkflowID    uuid.UUID
	CreatedAt     time.Time
	UpdatedAt     time.Time

	events []event.DomainEvent
}

type NewAssertionParams struct {
	Description   string
	Source        AssertionSource
	Path          string
	Operator      AssertionOperator
	ExpectedValue string
	StepID        uuid.UUID
	WorkflowID    uuid.UUID
	ProjectID     uuid.UUID
}

func NewAssertion(p NewAssertionParams) (*Assertion, error) {
	if err := ValidateShape(p.Source, p.Path, p.Operator, p.ExpectedValue); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	a := &Assertion{
		ID:            uuid.New(),
		Description:   strings.TrimSpace(p.Description),
		Source:        p.Source,
		Path:          normalizePath(p.Source, p.Path),
		Operator:      p.Operator,
		ExpectedValue: normalizeExpectedValue(p.Operator, p.ExpectedValue),
		StepID:        p.StepID,
		WorkflowID:    p.WorkflowID,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	a.recordEvent(AssertionCreated{
		ID:            uuid.New().String(),
		AssertionID:   a.ID.String(),
		WorkflowID:    a.WorkflowID.String(),
		StepID:        a.StepID.String(),
		ProjectID:     p.ProjectID.String(),
		Description:   a.Description,
		Source:        string(a.Source),
		Path:          a.Path,
		Operator:      string(a.Operator),
		ExpectedValue: a.ExpectedValue,
		Timestamp:     now,
	})
	return a, nil
}

type UpdateAssertionParams struct {
	Description   string
	Source        AssertionSource
	Path          string
	Operator      AssertionOperator
	ExpectedValue string
	ProjectID     uuid.UUID
}

func (a *Assertion) Update(p UpdateAssertionParams) error {
	if err := ValidateShape(p.Source, p.Path, p.Operator, p.ExpectedValue); err != nil {
		return err
	}

	a.Description = strings.TrimSpace(p.Description)
	a.Source = p.Source
	a.Path = normalizePath(p.Source, p.Path)
	a.Operator = p.Operator
	a.ExpectedValue = normalizeExpectedValue(p.Operator, p.ExpectedValue)
	a.UpdatedAt = time.Now().UTC()
	a.recordEvent(AssertionUpdated{
		ID:            uuid.New().String(),
		AssertionID:   a.ID.String(),
		WorkflowID:    a.WorkflowID.String(),
		StepID:        a.StepID.String(),
		ProjectID:     p.ProjectID.String(),
		Description:   a.Description,
		Source:        string(a.Source),
		Path:          a.Path,
		Operator:      string(a.Operator),
		ExpectedValue: a.ExpectedValue,
		Timestamp:     a.UpdatedAt,
	})
	return nil
}

func (a *Assertion) MarkDeleted(projectID uuid.UUID) {
	a.UpdatedAt = time.Now().UTC()
	a.recordEvent(AssertionDeleted{
		ID:            uuid.New().String(),
		AssertionID:   a.ID.String(),
		WorkflowID:    a.WorkflowID.String(),
		StepID:        a.StepID.String(),
		ProjectID:     projectID.String(),
		Source:        string(a.Source),
		Operator:      string(a.Operator),
		ExpectedValue: a.ExpectedValue,
		Timestamp:     a.UpdatedAt,
	})
}

func (a *Assertion) PullEvents() []event.DomainEvent {
	events := a.events
	a.events = nil
	return events
}

func (a *Assertion) recordEvent(evt event.DomainEvent) {
	a.events = append(a.events, evt)
}

func normalizePath(source AssertionSource, path string) string {
	if source == SourceStatus {
		return ""
	}
	return strings.TrimSpace(path)
}

func normalizeExpectedValue(operator AssertionOperator, expectedValue string) string {
	if !operator.RequiresExpectedValue() {
		return ""
	}
	return strings.TrimSpace(expectedValue)
}
