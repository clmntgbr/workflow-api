package di

import (
	workflowruncmd "go-api/internal/application/command/workflowrun"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	StartWorkflowRunHandler *workflowruncmd.StartWorkflowRunHandler
	WorkflowReadRepo        domainworkflow.WorkflowReadRepository
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	_ = env
	workflowWriteRepo := write.NewWorkflowWriteRepository(db)
	workflowReadRepo := read.NewWorkflowReadRepository(db)
	workflowRunWriteRepo := write.NewWorkflowRunWriteRepository(db)
	outboxRepo := outbox.NewRepository(db)

	return &Container{
		StartWorkflowRunHandler: workflowruncmd.NewStartWorkflowRunHandler(
			workflowWriteRepo,
			workflowRunWriteRepo,
			outboxRepo,
		),
		WorkflowReadRepo: workflowReadRepo,
	}
}
