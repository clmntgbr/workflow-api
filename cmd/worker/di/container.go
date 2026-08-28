package di

import (
	"log"

	eventassertion "go-api/internal/application/event/assertion"
	eventconnection "go-api/internal/application/event/connection"
	"go-api/internal/application/event/dedup"
	eventendpoint "go-api/internal/application/event/endpoint"
	eventinvoice "go-api/internal/application/event/invoice"
	eventproject "go-api/internal/application/event/project"
	eventstep "go-api/internal/application/event/step"
	eventsteprun "go-api/internal/application/event/steprun"
	eventsubscription "go-api/internal/application/event/subscription"
	eventuser "go-api/internal/application/event/user"
	eventvariable "go-api/internal/application/event/variable"
	eventworkflow "go-api/internal/application/event/workflow"
	eventworkflowrun "go-api/internal/application/event/workflowrun"
	"go-api/internal/application/registry"
	domainassertion "go-api/internal/domain/assertion"
	domainconnection "go-api/internal/domain/connection"
	domainendpoint "go-api/internal/domain/endpoint"
	domaininvoice "go-api/internal/domain/invoice"
	domainproject "go-api/internal/domain/project"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"
	domainvariable "go-api/internal/domain/variable"
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
	projectReadRepo := read.NewProjectReadRepository(db)
	workflowReadRepo := read.NewWorkflowReadRepository(db)
	stepReadRepo := read.NewStepReadRepository(db)
	connReadRepo := read.NewConnectionReadRepository(db)
	workflowRunWriteRepo := write.NewWorkflowRunWriteRepository(db)
	stepRunWriteRepo := write.NewStepRunWriteRepository(db)
	variableReadRepo := read.NewVariableReadRepository(db)
	assertionReadRepo := read.NewAssertionReadRepository(db)
	orchestrator := eventworkflowrun.NewOrchestrator(
		workflowRunWriteRepo,
		stepRunWriteRepo,
		stepReadRepo,
		connReadRepo,
		variableReadRepo,
		assertionReadRepo,
		outboxRepo,
	)
	enqueueStepRun := eventworkflowrun.NewEnqueueStepRunHandler(stepRunExecutor)
	publishUserRealtime := eventuser.NewPublishRealtimeHandler(realtimePublisher)
	publishProjectRealtime := eventproject.NewPublishRealtimeHandler(realtimePublisher)
	publishWorkflowRealtime := eventworkflow.NewPublishRealtimeHandler(realtimePublisher, projectReadRepo)
	publishEndpointRealtime := eventendpoint.NewPublishRealtimeHandler(realtimePublisher, projectReadRepo)
	publishStepRealtime := eventstep.NewPublishRealtimeHandler(realtimePublisher, projectReadRepo)
	publishConnectionRealtime := eventconnection.NewPublishRealtimeHandler(realtimePublisher, projectReadRepo)
	publishVariableRealtime := eventvariable.NewPublishRealtimeHandler(realtimePublisher, projectReadRepo)
	publishAssertionRealtime := eventassertion.NewPublishRealtimeHandler(realtimePublisher, projectReadRepo)
	publishWorkflowRunRealtime := eventworkflowrun.NewPublishRealtimeHandler(realtimePublisher, workflowReadRepo, projectReadRepo)
	publishStepRunRealtime := eventsteprun.NewPublishRealtimeHandler(realtimePublisher, projectReadRepo)
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
	reg.Register(domainuser.EventTypeUserActiveProjectChanged, dedup.With(
		dedupRepo,
		"user_active_project_changed",
		eventuser.NewUserActiveProjectChangedHandler().Handle,
	))
	reg.Register(domainuser.EventTypeUserActiveProjectChanged, dedup.With(
		dedupRepo,
		"publish_user_active_project_changed_realtime",
		publishUserRealtime.OnActiveProjectChanged,
	))

	reg.Register(domainproject.EventTypeProjectCreated, dedup.With(
		dedupRepo,
		"project_created",
		eventproject.NewProjectCreatedHandler().Handle,
	))
	reg.Register(domainproject.EventTypeProjectCreated, dedup.With(
		dedupRepo,
		"publish_project_created_realtime",
		publishProjectRealtime.OnCreated,
	))
	reg.Register(domainproject.EventTypeProjectUpdated, dedup.With(
		dedupRepo,
		"project_updated",
		eventproject.NewProjectUpdatedHandler().Handle,
	))
	reg.Register(domainproject.EventTypeProjectUpdated, dedup.With(
		dedupRepo,
		"publish_project_updated_realtime",
		publishProjectRealtime.OnUpdated,
	))
	reg.Register(domainproject.EventTypeProjectDeleted, dedup.With(
		dedupRepo,
		"project_deleted",
		eventproject.NewProjectDeletedHandler().Handle,
	))
	reg.Register(domainproject.EventTypeProjectDeleted, dedup.With(
		dedupRepo,
		"publish_project_deleted_realtime",
		publishProjectRealtime.OnDeleted,
	))
	reg.Register(domainproject.EventTypeProjectMemberAdded, dedup.With(
		dedupRepo,
		"project_member_added",
		eventproject.NewProjectMemberAddedHandler().Handle,
	))
	reg.Register(domainproject.EventTypeProjectMemberAdded, dedup.With(
		dedupRepo,
		"publish_project_member_added_realtime",
		publishProjectRealtime.OnMemberAdded,
	))
	reg.Register(domainproject.EventTypeProjectMemberRemoved, dedup.With(
		dedupRepo,
		"project_member_removed",
		eventproject.NewProjectMemberRemovedHandler().Handle,
	))
	reg.Register(domainproject.EventTypeProjectMemberRemoved, dedup.With(
		dedupRepo,
		"publish_project_member_removed_realtime",
		publishProjectRealtime.OnMemberRemoved,
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
	reg.Register(domainendpoint.EventTypeEndpointImported, dedup.With(
		dedupRepo,
		"endpoint_imported",
		eventendpoint.NewEndpointImportedHandler().Handle,
	))
	reg.Register(domainendpoint.EventTypeEndpointImported, dedup.With(
		dedupRepo,
		"publish_endpoint_imported_realtime",
		publishEndpointRealtime.OnImported,
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

	reg.Register(domainvariable.EventTypeVariableCreated, dedup.With(
		dedupRepo,
		"variable_created",
		eventvariable.NewVariableCreatedHandler().Handle,
	))
	reg.Register(domainvariable.EventTypeVariableCreated, dedup.With(
		dedupRepo,
		"publish_variable_created_realtime",
		publishVariableRealtime.OnCreated,
	))
	reg.Register(domainvariable.EventTypeVariableUpdated, dedup.With(
		dedupRepo,
		"variable_updated",
		eventvariable.NewVariableUpdatedHandler().Handle,
	))
	reg.Register(domainvariable.EventTypeVariableUpdated, dedup.With(
		dedupRepo,
		"publish_variable_updated_realtime",
		publishVariableRealtime.OnUpdated,
	))

	reg.Register(domainassertion.EventTypeAssertionCreated, dedup.With(
		dedupRepo,
		"assertion_created",
		eventassertion.NewAssertionCreatedHandler().Handle,
	))
	reg.Register(domainassertion.EventTypeAssertionCreated, dedup.With(
		dedupRepo,
		"publish_assertion_created_realtime",
		publishAssertionRealtime.OnCreated,
	))
	reg.Register(domainassertion.EventTypeAssertionUpdated, dedup.With(
		dedupRepo,
		"assertion_updated",
		eventassertion.NewAssertionUpdatedHandler().Handle,
	))
	reg.Register(domainassertion.EventTypeAssertionUpdated, dedup.With(
		dedupRepo,
		"publish_assertion_updated_realtime",
		publishAssertionRealtime.OnUpdated,
	))

	reg.Register(domainsubscription.EventTypeSubscriptionCreated, dedup.With(
		dedupRepo,
		"subscription_created",
		eventsubscription.NewSubscriptionCreatedHandler().Handle,
	))
	reg.Register(domainsubscription.EventTypeSubscriptionUpdated, dedup.With(
		dedupRepo,
		"subscription_updated",
		eventsubscription.NewSubscriptionUpdatedHandler().Handle,
	))

	reg.Register(domaininvoice.EventTypeInvoiceCreated, dedup.With(
		dedupRepo,
		"invoice_created",
		eventinvoice.NewInvoiceCreatedHandler().Handle,
	))
	reg.Register(domaininvoice.EventTypeInvoiceUpdated, dedup.With(
		dedupRepo,
		"invoice_updated",
		eventinvoice.NewInvoiceUpdatedHandler().Handle,
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
	reg.Register(domainworkflowrun.EventTypeWorkflowRunFinished, dedup.With(
		dedupRepo,
		"publish_workflow_run_finished_realtime",
		publishWorkflowRunRealtime.OnFinished,
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
