package quota

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type QuotaWriteRepository interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
	Save(ctx context.Context, quota *Quota) error
	Update(ctx context.Context, quota *Quota) error
	GetByID(ctx context.Context, id uuid.UUID) (*Quota, error)
}

type QuotaReadRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*QuotaView, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]QuotaView, error)
}

type QuotaView struct {
	ID   uuid.UUID
	Name string

	MaxProjectMembers  int
	MaxProjects             int
	MaxWorkflows            int
	MaxStepsPerWorkflow     int
	MaxEndpoints            int
	MaxVariablesPerWorkflow int

	MaxWorkflowRunsPerMonth    int
	MaxConcurrentRuns          int
	MinScheduleIntervalMinutes int

	RunHistoryRetentionDays int

	MaxStepTimeoutSeconds int
	MaxRetryCountPerStep  int
	MaxRequestBodySizeKB  int
	MaxResponseBodySizeKB int

	AllowsOpenAPIImport bool
	AllowsInsights      bool
	AllowsDataExport    bool
	ExecutorPriority    int

	CreatedAt time.Time
	UpdatedAt time.Time
}
