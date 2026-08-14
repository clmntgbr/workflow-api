package di

import (
	"log"

	eventconnection "go-api/internal/application/event/connection"
	"go-api/internal/application/event/dedup"
	eventendpoint "go-api/internal/application/event/endpoint"
	eventorganization "go-api/internal/application/event/organization"
	eventstep "go-api/internal/application/event/step"
	eventsteprun "go-api/internal/application/event/steprun"
	eventuser "go-api/internal/application/event/user"
	eventworkflow "go-api/internal/application/event/workflow"
	eventworkflowrun "go-api/internal/application/event/workflowrun"
	"go-api/internal/application/registry"
	domainconnection "go-api/internal/domain/connection"
	domainendpoint "go-api/internal/domain/endpoint"
	domainorganization "go-api/internal/domain/organization"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainuser "go-api/internal/domain/user"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"
	"go-api/internal/infrastructure/centrifugo"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/messaging/rabbitmq"
	"go-api/internal/infrastructure/notification"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/processed"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"

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

	executorTopology := rabbitmq.DefaultTopology(
		env.RabbitMQExecutorExchange,
		env.RabbitMQExecutorQueue,
		env.RabbitMQExecutorRoutingKey,
		env.RabbitMQRetryTTLMS,
	)

	conn, err := rabbitmq.Connect(env.RabbitMQURL, topology, executorTopology)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	publisher := rabbitmq.NewPublisher(conn, env.RabbitMQExchange)
	executorPublisher := rabbitmq.NewPublisher(conn, env.RabbitMQExecutorExchange)
	stepRunExecutor := rabbitmq.NewStepRunExecutor(executorPublisher)
	outboxRepo := outbox.NewRepository(db)
	relay := outbox.NewRelay(outboxRepo, publisher, env.OutboxPollInterval, 50)

	dedupRepo := processed.NewRepository(db)
	notifier := notification.NewLogNotifier()
	realtimePublisher := centrifugo.NewPublisher(env)
	orgReadRepo := read.NewOrganizationReadRepository(db)
	workflowReadRepo := read.NewWorkflowReadRepository(db)
	stepReadRepo := read.NewStepReadRepository(db)
	connReadRepo := read.NewConnectionReadRepository(db)
	workflowRunWriteRepo := write.NewWorkflowRunWriteRepository(db)
	stepRunWriteRepo := write.NewStepRunWriteRepository(db)
	stepRunReadRepo := read.NewStepRunReadRepository(db)
	variableWriteRepo := write.NewVariableWriteRepository(db)
	variableReadRepo := read.NewVariableReadRepository(db)
	orchestrator := eventworkflowrun.NewOrchestrator(
		workflowRunWriteRepo,
		stepRunWriteRepo,
		stepReadRepo,
		connReadRepo,
		variableWriteRepo,
		variableReadRepo,
		outboxRepo,
	)
	enqueueStepRun := eventworkflowrun.NewEnqueueStepRunHandler(stepRunExecutor)
	publishUserRealtime := eventuser.NewPublishRealtimeHandler(realtimePublisher)
	publishOrgRealtime := eventorganization.NewPublishRealtimeHandler(realtimePublisher)
	publishWorkflowRealtime := eventworkflow.NewPublishRealtimeHandler(realtimePublisher, orgReadRepo)
	publishEndpointRealtime := eventendpoint.NewPublishRealtimeHandler(realtimePublisher, orgReadRepo)
	publishStepRealtime := eventstep.NewPublishRealtimeHandler(realtimePublisher, orgReadRepo)
	publishConnectionRealtime := eventconnection.NewPublishRealtimeHandler(realtimePublisher, orgReadRepo)
	publishWorkflowRunRealtime := eventworkflowrun.NewPublishRealtimeHandler(realtimePublisher, workflowReadRepo, orgReadRepo)
	publishStepRunRealtime := eventsteprun.NewPublishRealtimeHandler(realtimePublisher, orgReadRepo, stepRunReadRepo)
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
	reg.Register(domainorganization.EventTypeOrganizationUpdated, dedup.With(
		dedupRepo,
		"publish_organization_updated_realtime",
		publishOrgRealtime.OnUpdated,
	))
	reg.Register(domainorganization.EventTypeOrganizationDeleted, dedup.With(
		dedupRepo,
		"organization_deleted",
		eventorganization.NewOrganizationDeletedHandler().Handle,
	))
	reg.Register(domainorganization.EventTypeOrganizationDeleted, dedup.With(
		dedupRepo,
		"publish_organization_deleted_realtime",
		publishOrgRealtime.OnDeleted,
	))
	reg.Register(domainorganization.EventTypeOrganizationMemberAdded, dedup.With(
		dedupRepo,
		"organization_member_added",
		eventorganization.NewOrganizationMemberAddedHandler().Handle,
	))
	reg.Register(domainorganization.EventTypeOrganizationMemberAdded, dedup.With(
		dedupRepo,
		"publish_organization_member_added_realtime",
		publishOrgRealtime.OnMemberAdded,
	))
	reg.Register(domainorganization.EventTypeOrganizationMemberRemoved, dedup.With(
		dedupRepo,
		"organization_member_removed",
		eventorganization.NewOrganizationMemberRemovedHandler().Handle,
	))
	reg.Register(domainorganization.EventTypeOrganizationMemberRemoved, dedup.With(
		dedupRepo,
		"publish_organization_member_removed_realtime",
		publishOrgRealtime.OnMemberRemoved,
	))

	reg.Register(domainworkflow.EventTypeWorkflowCreated, dedup.With(
		dedupRepo,
		"workflow_created",
		eventworkflow.NewWorkflowCreatedHandler().Handle,
	))
	reg.Register(domainworkflow.EventTypeWorkflowCreated, dedup.With(
		dedupRepo,
		"publish_workflow_created_realtime",
		publishWorkflowRealtime.OnCreated,
	))
	reg.Register(domainworkflow.EventTypeWorkflowUpdated, dedup.With(
		dedupRepo,
		"workflow_updated",
		eventworkflow.NewWorkflowUpdatedHandler().Handle,
	))
	reg.Register(domainworkflow.EventTypeWorkflowUpdated, dedup.With(
		dedupRepo,
		"publish_workflow_updated_realtime",
		publishWorkflowRealtime.OnUpdated,
	))
	reg.Register(domainworkflow.EventTypeWorkflowDeleted, dedup.With(
		dedupRepo,
		"workflow_deleted",
		eventworkflow.NewWorkflowDeletedHandler().Handle,
	))

	reg.Register(domainendpoint.EventTypeEndpointCreated, dedup.With(
		dedupRepo,
		"endpoint_created",
		eventendpoint.NewEndpointCreatedHandler().Handle,
	))
	reg.Register(domainendpoint.EventTypeEndpointCreated, dedup.With(
		dedupRepo,
		"publish_endpoint_created_realtime",
		publishEndpointRealtime.OnCreated,
	))
	reg.Register(domainendpoint.EventTypeEndpointUpdated, dedup.With(
		dedupRepo,
		"endpoint_updated",
		eventendpoint.NewEndpointUpdatedHandler().Handle,
	))
	reg.Register(domainendpoint.EventTypeEndpointUpdated, dedup.With(
		dedupRepo,
		"publish_endpoint_updated_realtime",
		publishEndpointRealtime.OnUpdated,
	))
	reg.Register(domainendpoint.EventTypeEndpointDeleted, dedup.With(
		dedupRepo,
		"endpoint_deleted",
		eventendpoint.NewEndpointDeletedHandler().Handle,
	))
	reg.Register(domainendpoint.EventTypeEndpointDeleted, dedup.With(
		dedupRepo,
		"publish_endpoint_deleted_realtime",
		publishEndpointRealtime.OnDeleted,
	))

	reg.Register(domainstep.EventTypeStepCreated, dedup.With(
		dedupRepo,
		"step_created",
		eventstep.NewStepCreatedHandler().Handle,
	))
	reg.Register(domainstep.EventTypeStepCreated, dedup.With(
		dedupRepo,
		"publish_step_created_realtime",
		publishStepRealtime.OnCreated,
	))
	reg.Register(domainstep.EventTypeStepUpdated, dedup.With(
		dedupRepo,
		"step_updated",
		eventstep.NewStepUpdatedHandler().Handle,
	))
	reg.Register(domainstep.EventTypeStepUpdated, dedup.With(
		dedupRepo,
		"publish_step_updated_realtime",
		publishStepRealtime.OnUpdated,
	))
	reg.Register(domainstep.EventTypeStepDeleted, dedup.With(
		dedupRepo,
		"step_deleted",
		eventstep.NewStepDeletedHandler().Handle,
	))
	reg.Register(domainstep.EventTypeStepDeleted, dedup.With(
		dedupRepo,
		"publish_step_deleted_realtime",
		publishStepRealtime.OnDeleted,
	))

	reg.Register(domainconnection.EventTypeConnectionCreated, dedup.With(
		dedupRepo,
		"connection_created",
		eventconnection.NewConnectionCreatedHandler().Handle,
	))
	reg.Register(domainconnection.EventTypeConnectionCreated, dedup.With(
		dedupRepo,
		"publish_connection_created_realtime",
		publishConnectionRealtime.OnCreated,
	))
	reg.Register(domainconnection.EventTypeConnectionDeleted, dedup.With(
		dedupRepo,
		"connection_deleted",
		eventconnection.NewConnectionDeletedHandler().Handle,
	))
	reg.Register(domainconnection.EventTypeConnectionDeleted, dedup.With(
		dedupRepo,
		"publish_connection_deleted_realtime",
		publishConnectionRealtime.OnDeleted,
	))

	reg.Register(domainworkflowrun.EventTypeWorkflowRunStarted, dedup.With(
		dedupRepo,
		"orchestrate_workflow_run_started",
		orchestrator.OnStarted,
	))
	reg.Register(domainworkflowrun.EventTypeWorkflowRunStarted, dedup.With(
		dedupRepo,
		"publish_workflow_run_started_realtime",
		publishWorkflowRunRealtime.OnStarted,
	))
	reg.Register(domainworkflowrun.EventTypeWorkflowRunSucceeded, dedup.With(
		dedupRepo,
		"publish_workflow_run_succeeded_realtime",
		publishWorkflowRunRealtime.OnSucceeded,
	))
	reg.Register(domainworkflowrun.EventTypeWorkflowRunFailed, dedup.With(
		dedupRepo,
		"publish_workflow_run_failed_realtime",
		publishWorkflowRunRealtime.OnFailed,
	))
	reg.Register(domainworkflowrun.EventTypeWorkflowRunCancelled, dedup.With(
		dedupRepo,
		"publish_workflow_run_cancelled_realtime",
		publishWorkflowRunRealtime.OnCancelled,
	))
	reg.Register(domainworkflowrun.EventTypeWorkflowRunScheduledSkipped, dedup.With(
		dedupRepo,
		"workflow_run_scheduled_skipped",
		eventworkflowrun.NewScheduledSkippedHandler().Handle,
	))

	reg.Register(domainsteprun.EventTypeStepRunQueued, dedup.With(
		dedupRepo,
		"enqueue_step_run_execute",
		enqueueStepRun.Handle,
	))
	reg.Register(domainsteprun.EventTypeStepRunStarted, dedup.With(
		dedupRepo,
		"publish_step_run_started_realtime",
		publishStepRunRealtime.OnStarted,
	))
	reg.Register(domainsteprun.EventTypeStepRunSucceeded, dedup.With(
		dedupRepo,
		"orchestrate_step_run_succeeded",
		orchestrator.OnSucceeded,
	))
	reg.Register(domainsteprun.EventTypeStepRunSucceeded, dedup.With(
		dedupRepo,
		"publish_step_run_succeeded_realtime",
		publishStepRunRealtime.OnSucceeded,
	))
	reg.Register(domainsteprun.EventTypeStepRunFailed, dedup.With(
		dedupRepo,
		"orchestrate_step_run_failed",
		orchestrator.OnFailed,
	))
	reg.Register(domainsteprun.EventTypeStepRunFailed, dedup.With(
		dedupRepo,
		"publish_step_run_failed_realtime",
		publishStepRunRealtime.OnFailed,
	))

	consumer := rabbitmq.NewConsumer(conn, reg, env.WorkerConcurrency, env.WorkerMaxRetries)

	return &Container{
		Relay:    relay,
		Consumer: consumer,
		Conn:     conn,
	}
}
