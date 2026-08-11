package di

import (
	"log"

	"go-api/internal/application/event/dedup"
	eventorganization "go-api/internal/application/event/organization"
	eventuser "go-api/internal/application/event/user"
	eventworkflow "go-api/internal/application/event/workflow"
	"go-api/internal/application/registry"
	domainorganization "go-api/internal/domain/organization"
	domainuser "go-api/internal/domain/user"
	domainworkflow "go-api/internal/domain/workflow"
	"go-api/internal/infrastructure/centrifugo"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/notification"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"

	"gorm.io/gorm"
)

type Container struct {
	Relay    *outbox.Relay
	Consumer *rabbitmq.Consumer
	Conn     *rabbitmq.Connection
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	topology := rabbitmq.DefaultTopology(
		env.RabbitMQExchange,
		env.RabbitMQQueue,
		env.RabbitMQRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	publisher := rabbitmq.NewPublisher(conn, env.RabbitMQExchange)
	outboxRepo := outbox.NewRepository(db)
	relay := outbox.NewRelay(outboxRepo, publisher, env.OutboxPollInterval, 50)

	dedupRepo := processed.NewRepository(db)
	notifier := notification.NewLogNotifier()
	realtimePublisher := centrifugo.NewPublisher(env)
	publishUserRealtime := eventuser.NewPublishRealtimeHandler(realtimePublisher)
	publishOrgRealtime := eventorganization.NewPublishRealtimeHandler(realtimePublisher)
	reg := registry.NewHandlerRegistry()

	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"user_created",
		eventuser.NewUserCreatedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"notify_user_on_created",
		eventuser.NewNotifyUserOnCreatedHandler(notifier).Handle,
	))
	reg.Register(domainuser.EventTypeUserCreated, dedup.With(
		dedupRepo,
		"publish_user_created_realtime",
		publishUserRealtime.OnCreated,
	))
	reg.Register(domainuser.EventTypeUserUpdated, dedup.With(
		dedupRepo,
		"user_updated",
		eventuser.NewUserUpdatedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserUpdated, dedup.With(
		dedupRepo,
		"publish_user_updated_realtime",
		publishUserRealtime.OnUpdated,
	))
	reg.Register(domainuser.EventTypeUserDeleted, dedup.With(
		dedupRepo,
		"user_deleted",
		eventuser.NewUserDeletedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserDeleted, dedup.With(
		dedupRepo,
		"publish_user_deleted_realtime",
		publishUserRealtime.OnDeleted,
	))
	reg.Register(domainuser.EventTypeUserActiveOrganizationChanged, dedup.With(
		dedupRepo,
		"user_active_organization_changed",
		eventuser.NewUserActiveOrganizationChangedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserActiveOrganizationChanged, dedup.With(
		dedupRepo,
		"publish_user_active_organization_changed_realtime",
		publishUserRealtime.OnActiveOrganizationChanged,
	))

	reg.Register(domainorganization.EventTypeOrganizationCreated, dedup.With(
		dedupRepo,
		"organization_created",
		eventorganization.NewOrganizationCreatedHandler().Handle,
	))
	reg.Register(domainorganization.EventTypeOrganizationCreated, dedup.With(
		dedupRepo,
		"publish_organization_created_realtime",
		publishOrgRealtime.OnCreated,
	))
	reg.Register(domainorganization.EventTypeOrganizationUpdated, dedup.With(
		dedupRepo,
		"organization_updated",
		eventorganization.NewOrganizationUpdatedHandler().Handle,
	))
	reg.Register(domainorganization.EventTypeOrganizationDeleted, dedup.With(
		dedupRepo,
		"organization_deleted",
		eventorganization.NewOrganizationDeletedHandler().Handle,
	))
	reg.Register(domainorganization.EventTypeOrganizationMemberAdded, dedup.With(
		dedupRepo,
		"organization_member_added",
		eventorganization.NewOrganizationMemberAddedHandler().Handle,
	))
	reg.Register(domainorganization.EventTypeOrganizationMemberRemoved, dedup.With(
		dedupRepo,
		"organization_member_removed",
		eventorganization.NewOrganizationMemberRemovedHandler().Handle,
	))

	reg.Register(domainworkflow.EventTypeWorkflowCreated, dedup.With(
		dedupRepo,
		"workflow_created",
		eventworkflow.NewWorkflowCreatedHandler().Handle,
	))
	reg.Register(domainworkflow.EventTypeWorkflowUpdated, dedup.With(
		dedupRepo,
		"workflow_updated",
		eventworkflow.NewWorkflowUpdatedHandler().Handle,
	))
	reg.Register(domainworkflow.EventTypeWorkflowDeleted, dedup.With(
		dedupRepo,
		"workflow_deleted",
		eventworkflow.NewWorkflowDeletedHandler().Handle,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Relay:    relay,
		Consumer: consumer,
		Conn:     conn,
	}
}
