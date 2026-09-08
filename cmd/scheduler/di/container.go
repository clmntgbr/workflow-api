package di

import (
	cmdquota "go-api/internal/application/command/quota"
	stepruncmd "go-api/internal/application/command/steprun"
	workflowcmd "go-api/internal/application/command/workflow"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	StartWorkflowRunHandler  *workflowruncmd.StartWorkflowRunHandler
	ClaimDueWorkflowsHandler *workflowcmd.ClaimDueWorkflowsHandler
	FailStaleStepRunsHandler *stepruncmd.FailStaleStepRunsHandler
	BatchSize                int
	Concurrency              int
	MaxBatchesPerTick        int
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	workflowWriteRepo := write.NewWorkflowWriteRepository(db)
	workflowRunWriteRepo := write.NewWorkflowRunWriteRepository(db)
	stepRunWriteRepo := write.NewStepRunWriteRepository(db)
	variableReadRepo := read.NewVariableReadRepository(db)
	connReadRepo := read.NewConnectionReadRepository(db)
	stepRunReadRepo := read.NewStepRunReadRepository(db)
	outboxRepo := outbox.NewRepository(db)
	userReadRepo := read.NewUserReadRepository(db)
	projectReadRepo := read.NewProjectReadRepository(db)
	workflowReadRepo := read.NewWorkflowReadRepository(db)
	endpointReadRepo := read.NewEndpointReadRepository(db)
	quotaReadRepo := read.NewQuotaReadRepository(db)
	planReadRepo := read.NewPlanReadRepository(db, quotaReadRepo)
	subscriptionReadRepo := read.NewSubscriptionReadRepository(db, planReadRepo)
	workflowRunReadRepo := read.NewWorkflowRunReadRepository(db)
	stepReadRepo := read.NewStepReadRepository(db)
	assertionReadRepo := read.NewAssertionReadRepository(db)

	getQuotaUsageHandler := querysubscription.NewGetQuotaUsageHandler(
		userReadRepo,
		subscriptionReadRepo,
		projectReadRepo,
		workflowReadRepo,
		endpointReadRepo,
		workflowRunReadRepo,
	)
	assertCreateAllowedHandler := cmdquota.NewAssertCreateAllowedHandler(
		getQuotaUsageHandler,
		stepReadRepo,
		variableReadRepo,
		assertionReadRepo,
		projectReadRepo,
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
			stepReadRepo,
			connReadRepo,
			stepRunReadRepo,
			stepRunWriteRepo,
			outboxRepo,
			assertCreateAllowedHandler,
		),
		ClaimDueWorkflowsHandler: workflowcmd.NewClaimDueWorkflowsHandler(
			workflowWriteRepo,
			outboxRepo,
		),
		FailStaleStepRunsHandler: stepruncmd.NewFailStaleStepRunsHandler(
			stepRunWriteRepo,
			workflowRunWriteRepo,
			outboxRepo,
		),
		BatchSize:         batchSize,
		Concurrency:       concurrency,
		MaxBatchesPerTick: maxBatches,
	}
}
