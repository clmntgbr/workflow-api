package di

import (
	"log"

	cmdquota "go-api/internal/application/command/quota"
	insightcmd "go-api/internal/application/command/insight"
	stepruncmd "go-api/internal/application/command/steprun"
	eventsteprun "go-api/internal/application/event/steprun"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/application/registry"
	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/httpexecutor"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.ExecutorTopology(
		env.RabbitMQExecutorExchange,
		env.RabbitMQExecutorQueue,
		env.RabbitMQExecutorRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	stepRunRepo := write.NewStepRunWriteRepository(db)
	insightRepo := write.NewInsightWriteRepository(db)
	outboxRepo := outbox.NewRepository(db)
	httpClient := httpexecutor.New()

	userReadRepo := read.NewUserReadRepository(db)
	projectReadRepo := read.NewProjectReadRepository(db)
	workflowReadRepo := read.NewWorkflowReadRepository(db)
	endpointReadRepo := read.NewEndpointReadRepository(db)
	workflowRunReadRepo := read.NewWorkflowRunReadRepository(db)
	stepReadRepo := read.NewStepReadRepository(db)
	variableReadRepo := read.NewVariableReadRepository(db)
	assertionReadRepo := read.NewAssertionReadRepository(db)
	quotaReadRepo := read.NewQuotaReadRepository(db)
	planReadRepo := read.NewPlanReadRepository(db, quotaReadRepo)
	subscriptionReadRepo := read.NewSubscriptionReadRepository(db, planReadRepo)

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

	startHandler := stepruncmd.NewStartStepRunHandler(stepRunRepo, outboxRepo)
	succeedHandler := stepruncmd.NewSucceedStepRunHandler(stepRunRepo, outboxRepo)
	failHandler := stepruncmd.NewFailStepRunHandler(stepRunRepo, outboxRepo)
	incrementHandler := stepruncmd.NewIncrementStepRunAttemptHandler(stepRunRepo)
	createInsightHandler := insightcmd.NewCreateInsightHandler(insightRepo)

	executeHandler := eventsteprun.NewExecuteHandler(
		stepRunRepo,
		httpClient,
		startHandler,
		succeedHandler,
		failHandler,
		incrementHandler,
		createInsightHandler,
		assertCreateAllowedHandler,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(port.EventTypeStepRunExecute, executeHandler.Handle)

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Consumer: consumer,
		Conn:     conn,
	}
}
