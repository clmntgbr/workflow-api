package di

import (
	workflowruncmd "go-api/internal/application/command/workflowrun"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
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
	outboxRepo := outbox.NewRepository(db)

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
			outboxRepo,
		),
		WorkflowWriteRepo: workflowWriteRepo,
		BatchSize:         batchSize,
		Concurrency:       concurrency,
		MaxBatchesPerTick: maxBatches,
	}
}
