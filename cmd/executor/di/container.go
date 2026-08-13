package di

import (
	"log"

	eventsteprun "go-api/internal/application/event/steprun"
	"go-api/internal/application/registry"
	"go-api/internal/domain/port"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/messaging/rabbitmq"
)

type Container struct {
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(env *config.Config) *Container {
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

	reg := registry.NewHandlerRegistry()
	reg.Register(port.EventTypeStepRunExecute, eventsteprun.NewExecuteHandler().Handle)

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Consumer: consumer,
		Conn:     conn,
	}
}
