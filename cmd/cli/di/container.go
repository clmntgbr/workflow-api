package di

import (
	cmdquota "go-api/internal/application/command/quota"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	StartWorkflowRunHandler *workflowruncmd.StartWorkflowRunHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	_ = env
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

	return &Container{
		StartWorkflowRunHandler: workflowruncmd.NewStartWorkflowRunHandler(
			workflowWriteRepo,
			workflowRunWriteRepo,
			variableReadRepo,
			outboxRepo,
			assertCreateAllowedHandler,
		),
	}
}
