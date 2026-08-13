package di

import (
	"log"

	stepruncmd "go-api/internal/application/command/steprun"
	eventsteprun "go-api/internal/application/event/steprun"
	"go-api/internal/application/registry"
	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/httpexecutor"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/write"

	"gorm.io/gorm"
)

type Container struct {
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
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
	outboxRepo := outbox.NewRepository(db)
	httpClient := httpexecutor.New()

	startHandler := stepruncmd.NewStartStepRunHandler(stepRunRepo, outboxRepo)
	succeedHandler := stepruncmd.NewSucceedStepRunHandler(stepRunRepo, outboxRepo)
	failHandler := stepruncmd.NewFailStepRunHandler(stepRunRepo, outboxRepo)
	incrementHandler := stepruncmd.NewIncrementStepRunAttemptHandler(stepRunRepo)

	executeHandler := eventsteprun.NewExecuteHandler(
		stepRunRepo,
		httpClient,
		startHandler,
		succeedHandler,
		failHandler,
		incrementHandler,
	)

	reg := registry.NewHandlerRegistry()
	reg.Register(port.EventTypeStepRunExecute, executeHandler.Handle)

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Consumer: consumer,
		Conn:     conn,
	}
}
