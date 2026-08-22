package di

import (
	cmdquota "go-api/internal/application/command/quota"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	querysubscription "go-api/internal/application/query/subscription"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	StartWorkflowRunHandler *workflowruncmd.StartWorkflowRunHandler
	WorkflowWriteRepo       domainworkflow.WorkflowWriteRepository
	BatchSize               int
	Concurrency             int
	MaxBatchesPerTick       int
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	workflowWriteRepo := write.NewWorkflowWriteRepository(db)
	workflowRunWriteRepo := write.NewWorkflowRunWriteRepository(db)
	variableReadRepo := read.NewVariableReadRepository(db)
	outboxRepo := outbox.NewRepository(db)
	userReadRepo := read.NewUserReadRepository(db)
	orgReadRepo := read.NewOrganizationReadRepository(db)
	workflowReadRepo := read.NewWorkflowReadRepository(db)
	endpointReadRepo := read.NewEndpointReadRepository(db)
	quotaReadRepo := read.NewQuotaReadRepository(db)
	planReadRepo := read.NewPlanReadRepository(db, quotaReadRepo)
	subscriptionReadRepo := read.NewSubscriptionReadRepository(db, planReadRepo)
	workflowRunReadRepo := read.NewWorkflowRunReadRepository(db)
	stepReadRepo := read.NewStepReadRepository(db)

	getQuotaUsageHandler := querysubscription.NewGetQuotaUsageHandler(
		userReadRepo,
		subscriptionReadRepo,
		orgReadRepo,
		workflowReadRepo,
		endpointReadRepo,
		workflowRunReadRepo,
	)
	assertCreateAllowedHandler := cmdquota.NewAssertCreateAllowedHandler(
		getQuotaUsageHandler,
		stepReadRepo,
		orgReadRepo,
		userReadRepo,
	)

	batchSize := env.SchedulerBatchSize
	if batchSize <= 0 {
		batchSize = 100
	}
	concurrency := env.SchedulerConcurrency
	if concurrency <= 0 {
		concurrency = 32
	}
	maxBatches := env.SchedulerMaxBatchesPerTick
	if maxBatches <= 0 {
		maxBatches = 100
	}

	return &Container{
		StartWorkflowRunHandler: workflowruncmd.NewStartWorkflowRunHandler(
			workflowWriteRepo,
			workflowRunWriteRepo,
			variableReadRepo,
			outboxRepo,
			assertCreateAllowedHandler,
		),
		WorkflowWriteRepo: workflowWriteRepo,
		BatchSize:         batchSize,
		Concurrency:       concurrency,
		MaxBatchesPerTick: maxBatches,
	}
}
