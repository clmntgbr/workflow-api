package steprun

import (
	"context"
	"time"

	domainassertion "go-api/internal/domain/assertion"
	"go-api/internal/domain/httpquery"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type StepRunWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, run *StepRun) error
	Update(ctx context.Context, run *StepRun) error
	GetByID(ctx context.Context, id uuid.UUID) (*StepRun, error)
	FindByWorkflowRunID(ctx context.Context, workflowRunID uuid.UUID) ([]*StepRun, error)
	ClaimDueWaiting(ctx context.Context, now time.Time, limit int) ([]*StepRun, error)
}

type StepRunReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*StepRunView, error)
	FindByWorkflowRunID(ctx context.Context, workflowRunID uuid.UUID) ([]StepRunView, error)
	FindByWorkflowRunIDs(ctx context.Context, workflowRunIDs []uuid.UUID) ([]StepRunView, error)
	FindLatestCompletedByStepID(ctx context.Context, stepID uuid.UUID) (*StepRunView, error)
	FindLatestStatusByStepIDs(ctx context.Context, stepIDs []uuid.UUID) (map[uuid.UUID]Status, error)
	FindStatusByWorkflowRunIDAndStepIDs(
		ctx context.Context,
		workflowRunID uuid.UUID,
		stepIDs []uuid.UUID,
	) (map[uuid.UUID]Status, error)
}

type StepRunView struct {
	ID                 uuid.UUID
	WorkflowRunID      uuid.UUID
	StepID             uuid.UUID
	WorkflowID         uuid.UUID
	EndpointID           *uuid.UUID
	ProjectID            uuid.UUID
	StepType             domainstep.Type
	DelayDurationSeconds int
	Name                 string
	Description        string
	URL                string
	Method             string
	Headers            map[string]string
	Query              httpquery.Params
	Body               map[string]any
	Timeout            int
	RetryOnFailure     bool
	RetryCount         int
	RetryDelay         int
	Index              string
	ExecutionOrder     int
	TreeIndex          int
	Position           domainstep.Position
	Status             Status
	Attempt            int
	VariableExtracts   []VariableExtract
	Assertions         []domainassertion.Snapshot
	ResponseSnapshot   *ResponseSnapshot
	ExtractedVariables map[string]any
	AssertionsResult   []domainassertion.Result
	StartedAt          *time.Time
	FinishedAt         *time.Time
	ResumeAt           *time.Time
	Error              string
	CreatedAt          time.Time
	UpdatedAt          time.Time
}
