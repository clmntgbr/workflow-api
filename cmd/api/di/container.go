package di

import (
	"log"

	authcmd "go-api/internal/application/command/auth"
	conncmd "go-api/internal/application/command/connection"
	endpointcmd "go-api/internal/application/command/endpoint"
	identitycmd "go-api/internal/application/command/identity"
	cmdquota "go-api/internal/application/command/quota"
	projectcmd "go-api/internal/application/command/project"
	subscriptioncmd "go-api/internal/application/command/subscription"
	stepcmd "go-api/internal/application/command/step"
	usercmd "go-api/internal/application/command/user"
	variablecmd "go-api/internal/application/command/variable"
	workflowcmd "go-api/internal/application/command/workflow"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	queryconn "go-api/internal/application/query/connection"
	queryendpoint "go-api/internal/application/query/endpoint"
	queryinvoice "go-api/internal/application/query/invoice"
	queryinsight "go-api/internal/application/query/insight"
	queryproject "go-api/internal/application/query/project"
	queryplan "go-api/internal/application/query/plan"
	querysubscription "go-api/internal/application/query/subscription"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryuser "go-api/internal/application/query/user"
	queryvariable "go-api/internal/application/query/variable"
	queryworkflow "go-api/internal/application/query/workflow"
	queryworkflowrun "go-api/internal/application/query/workflowrun"
	infraClerk "go-api/internal/infrastructure/clerk"
	"go-api/internal/infrastructure/config"
	infraopenapi "go-api/internal/infrastructure/openapi"
	infrastripe "go-api/internal/infrastructure/stripe"
	"go-api/internal/infrastructure/persistence/outbox"
	"go-api/internal/infrastructure/persistence/read"
	"go-api/internal/infrastructure/persistence/write"
	httphandler "go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/middleware"

	"gorm.io/gorm"
)

type Container struct {
	AuthenticateMiddleware *middleware.AuthenticateMiddleware
	UserWebhookMiddleware  *middleware.UserWebhookMiddleware
	UserWebhookHandler     *httphandler.UserWebhookHandler
	BillingWebhookMiddleware *middleware.BillingWebhookMiddleware
	BillingWebhookHandler    *httphandler.BillingWebhookHandler
	UserHandler            *httphandler.UserHandler
	ProjectHandler    *httphandler.ProjectHandler
	WorkflowHandler        *httphandler.WorkflowHandler
	EndpointHandler        *httphandler.EndpointHandler
	StepHandler            *httphandler.StepHandler
	ConnectionHandler      *httphandler.ConnectionHandler
	WorkflowRunHandler     *httphandler.WorkflowRunHandler
	VariableHandler        *httphandler.VariableHandler
	PlanHandler            *httphandler.PlanHandler
	SubscriptionHandler    *httphandler.SubscriptionHandler
	InvoiceHandler         *httphandler.InvoiceHandler
	RealtimeHandler        *httphandler.RealtimeHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	jwksProvider, err := infraClerk.NewJWKSProvider(env)
	if err != nil {
		log.Fatalf("failed to create JWKS provider: %v", err)
	}

	userWriteRepo := write.NewUserWriteRepository(db)
	userReadRepo := read.NewUserReadRepository(db)
	projectWriteRepo := write.NewProjectWriteRepository(db)
	projectReadRepo := read.NewProjectReadRepository(db)
	workflowWriteRepo := write.NewWorkflowWriteRepository(db)
	workflowReadRepo := read.NewWorkflowReadRepository(db)
	endpointWriteRepo := write.NewEndpointWriteRepository(db)
	endpointReadRepo := read.NewEndpointReadRepository(db)
	stepWriteRepo := write.NewStepWriteRepository(db)
	stepReadRepo := read.NewStepReadRepository(db)
	connWriteRepo := write.NewConnectionWriteRepository(db)
	connReadRepo := read.NewConnectionReadRepository(db)
	workflowRunWriteRepo := write.NewWorkflowRunWriteRepository(db)
	workflowRunReadRepo := read.NewWorkflowRunReadRepository(db)
	stepRunWriteRepo := write.NewStepRunWriteRepository(db)
	stepRunReadRepo := read.NewStepRunReadRepository(db)
	insightReadRepo := read.NewInsightReadRepository(db)
	variableWriteRepo := write.NewVariableWriteRepository(db)
	variableReadRepo := read.NewVariableReadRepository(db)
	quotaReadRepo := read.NewQuotaReadRepository(db)
	planWriteRepo := write.NewPlanWriteRepository(db)
	planReadRepo := read.NewPlanReadRepository(db, quotaReadRepo)
	subscriptionWriteRepo := write.NewSubscriptionWriteRepository(db)
	invoiceWriteRepo := write.NewInvoiceWriteRepository(db)
	invoiceReadRepo := read.NewInvoiceReadRepository(db)
	outboxRepo := outbox.NewRepository(db)

	createUserHandler := usercmd.NewCreateUserHandler(
		userWriteRepo,
		projectWriteRepo,
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
	)
	updateUserHandler := usercmd.NewUpdateUserHandler(userWriteRepo, outboxRepo)
	getUserByExternalIDHandler := usercmd.NewGetUserByExternalIDHandler(userWriteRepo)
	deleteUserByExternalIDHandler := usercmd.NewDeleteUserByExternalIDHandler(userWriteRepo, outboxRepo)
	setActiveProjectHandler := usercmd.NewSetActiveProjectHandler(userWriteRepo, projectWriteRepo, outboxRepo)
	validateTokenHandler := authcmd.NewValidateTokenHandler(jwksProvider, userWriteRepo)
	fetchUserHandler := identitycmd.NewFetchUserHandler(infraClerk.NewUserGateway(env.ClerkSecretKey))
	getUserByIDHandler := queryuser.NewGetUserByIDHandler(userReadRepo)

	updateOrgHandler := projectcmd.NewUpdateProjectHandler(projectWriteRepo, outboxRepo)
	deleteOrgHandler := projectcmd.NewDeleteProjectHandler(projectWriteRepo, outboxRepo)
	removeOrgMemberHandler := projectcmd.NewRemoveProjectMemberHandler(projectWriteRepo, userWriteRepo, outboxRepo)
	getProjectByIDHandler := queryproject.NewGetProjectByIDHandler(projectReadRepo)
	listProjectsByUserHandler := queryproject.NewListProjectsByUserHandler(projectReadRepo)

	updateWorkflowHandler := workflowcmd.NewUpdateWorkflowHandler(workflowWriteRepo, outboxRepo)
	activateWorkflowHandler := workflowcmd.NewActivateWorkflowHandler(workflowWriteRepo, outboxRepo)
	deactivateWorkflowHandler := workflowcmd.NewDeactivateWorkflowHandler(workflowWriteRepo, outboxRepo)
	deleteWorkflowHandler := workflowcmd.NewDeleteWorkflowHandler(workflowWriteRepo, outboxRepo)
	getWorkflowByIDHandler := queryworkflow.NewGetWorkflowByIDHandler(workflowReadRepo)
	listWorkflowsByProjectHandler := queryworkflow.NewListWorkflowsByProjectHandler(workflowReadRepo)

	updateEndpointHandler := endpointcmd.NewUpdateEndpointHandler(endpointWriteRepo, outboxRepo)
	deleteEndpointHandler := endpointcmd.NewDeleteEndpointHandler(endpointWriteRepo, outboxRepo)
	getEndpointByIDHandler := queryendpoint.NewGetEndpointByIDHandler(endpointReadRepo)
	listEndpointsByOrgHandler := queryendpoint.NewListEndpointsByProjectHandler(endpointReadRepo)

	createConnHandler := conncmd.NewCreateConnectionHandler(
		connWriteRepo,
		connReadRepo,
		stepReadRepo,
		stepWriteRepo,
		outboxRepo,
	)
	deleteConnHandler := conncmd.NewDeleteConnectionHandler(
		connWriteRepo,
		connReadRepo,
		stepReadRepo,
		stepWriteRepo,
		outboxRepo,
	)
	listConnsByWorkflowHandler := queryconn.NewListConnectionsByWorkflowHandler(connReadRepo)

	updateStepHandler := stepcmd.NewUpdateStepHandler(stepWriteRepo, outboxRepo)
	updateStepPositionHandler := stepcmd.NewUpdateStepPositionHandler(
		stepWriteRepo,
		stepReadRepo,
		connReadRepo,
		outboxRepo,
	)
	deleteStepHandler := stepcmd.NewDeleteStepHandler(
		stepWriteRepo,
		stepReadRepo,
		connWriteRepo,
		connReadRepo,
		variableWriteRepo,
		outboxRepo,
	)
	getStepByIDHandler := querystep.NewGetStepByIDHandler(stepReadRepo)
	listStepsByWorkflowHandler := querystep.NewListStepsByWorkflowHandler(stepReadRepo)

	createVariableHandler := variablecmd.NewCreateVariableHandler(variableWriteRepo, variableReadRepo, stepWriteRepo, outboxRepo)
	updateVariableHandler := variablecmd.NewUpdateVariableHandler(variableWriteRepo, outboxRepo)
	deleteVariableHandler := variablecmd.NewDeleteVariableHandler(variableWriteRepo, stepReadRepo)
	getVariableByIDHandler := queryvariable.NewGetVariableByIDHandler(variableReadRepo)
	listVariablesByWorkflowHandler := queryvariable.NewListVariablesByWorkflowHandler(variableReadRepo)
	listAvailableVariablesHandler := queryvariable.NewListAvailableVariablesHandler(variableReadRepo, connReadRepo)
	searchVariablePathsHandler := queryvariable.NewSearchVariablePathsHandler(stepRunReadRepo)

	getWorkflowRunByIDHandler := queryworkflowrun.NewGetWorkflowRunByIDHandler(workflowRunReadRepo)
	getWorkflowRunAnalyticsHandler := queryworkflowrun.NewGetWorkflowRunAnalyticsHandler(workflowRunReadRepo)
	listWorkflowRunsHandler := queryworkflowrun.NewListWorkflowRunsByWorkflowHandler(workflowRunReadRepo)
	listWorkflowRunsByProjectHandler := queryworkflowrun.NewListWorkflowRunsByProjectHandler(workflowRunReadRepo)
	listStepRunsHandler := querysteprun.NewListStepRunsByWorkflowRunHandler(stepRunReadRepo)
	listStepRunsByIDsHandler := querysteprun.NewListStepRunsByWorkflowRunIDsHandler(stepRunReadRepo)
	latestStepRunStatusesHandler := querysteprun.NewGetLatestStepRunStatusesByStepIDsHandler(stepRunReadRepo)
	listInsightsByIDsHandler := queryinsight.NewListInsightsByStepRunIDsHandler(insightReadRepo)

	listActivePlansHandler := queryplan.NewListActivePlansHandler(planReadRepo)

	subscriptionReadRepo := read.NewSubscriptionReadRepository(db, planReadRepo)
	stripeSubscriptionGateway := infrastripe.NewSubscriptionGateway(env)
	stripeCheckoutGateway := infrastripe.NewCheckoutSessionGateway(env)
	stripeBillingPortalGateway := infrastripe.NewBillingPortalGateway(env)

	getCurrentSubscriptionHandler := querysubscription.NewGetCurrentSubscriptionHandler(userReadRepo, subscriptionReadRepo)
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
		projectReadRepo,
		userReadRepo,
	)

	createOrgHandler := projectcmd.NewCreateProjectHandler(
		projectWriteRepo,
		userWriteRepo,
		outboxRepo,
		assertCreateAllowedHandler,
	)

	startWorkflowRunHandler := workflowruncmd.NewStartWorkflowRunHandler(
		workflowWriteRepo,
		workflowRunWriteRepo,
		variableReadRepo,
		outboxRepo,
		assertCreateAllowedHandler,
	)
	cancelWorkflowRunHandler := workflowruncmd.NewCancelWorkflowRunHandler(
		workflowWriteRepo,
		workflowRunWriteRepo,
		stepRunWriteRepo,
		outboxRepo,
	)
	createWorkflowHandler := workflowcmd.NewCreateWorkflowHandler(
		workflowWriteRepo,
		outboxRepo,
		assertCreateAllowedHandler,
	)
	createEndpointHandler := endpointcmd.NewCreateEndpointHandler(
		endpointWriteRepo,
		outboxRepo,
		assertCreateAllowedHandler,
	)
	importEndpointsHandler := endpointcmd.NewImportEndpointsFromOpenAPIHandler(
		endpointWriteRepo,
		infraopenapi.NewParser(),
		outboxRepo,
		assertCreateAllowedHandler,
	)
	createStepHandler := stepcmd.NewCreateStepHandler(
		stepWriteRepo,
		stepReadRepo,
		connReadRepo,
		endpointReadRepo,
		workflowReadRepo,
		outboxRepo,
		assertCreateAllowedHandler,
	)

	previewPlanChangeHandler := querysubscription.NewPreviewPlanChangeHandler(
		userReadRepo,
		planReadRepo,
		subscriptionReadRepo,
		stripeSubscriptionGateway,
	)
	createSubscriptionHandler := subscriptioncmd.NewCreateSubscriptionHandler(
		userReadRepo,
		planReadRepo,
		subscriptionReadRepo,
		subscriptionWriteRepo,
		outboxRepo,
		fetchUserHandler,
		stripeCheckoutGateway,
		stripeSubscriptionGateway,
	)
	createBillingPortalHandler := subscriptioncmd.NewCreateBillingPortalHandler(
		userReadRepo,
		subscriptionReadRepo,
		stripeBillingPortalGateway,
	)

	checkoutCompletedHandler := subscriptioncmd.NewCheckoutCompletedHandler(
		userWriteRepo,
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
		stripeSubscriptionGateway,
	)
	subscriptionUpdatedHandler := subscriptioncmd.NewSubscriptionUpdatedHandler(
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
	)
	subscriptionDeletedHandler := subscriptioncmd.NewSubscriptionDeletedHandler(
		planWriteRepo,
		subscriptionWriteRepo,
		outboxRepo,
	)
	invoicePaymentSucceededHandler := subscriptioncmd.NewInvoicePaymentSucceededHandler(
		subscriptionWriteRepo,
		outboxRepo,
	)
	invoicePaymentFailedHandler := subscriptioncmd.NewInvoicePaymentFailedHandler(
		subscriptionWriteRepo,
		outboxRepo,
	)
	upsertInvoiceHandler := subscriptioncmd.NewUpsertInvoiceHandler(
		invoiceWriteRepo,
		subscriptionWriteRepo,
		userWriteRepo,
		outboxRepo,
	)
	listInvoicesHandler := queryinvoice.NewListInvoicesHandler(invoiceReadRepo)

	return &Container{
		AuthenticateMiddleware: middleware.NewAuthenticateMiddleware(
			validateTokenHandler,
			fetchUserHandler,
			createUserHandler,
		),
		UserWebhookMiddleware: middleware.NewUserWebhookMiddleware(env.ClerkWebhookSecret),
		UserWebhookHandler: httphandler.NewUserWebhookHandler(
			getUserByExternalIDHandler,
			createUserHandler,
			updateUserHandler,
			deleteUserByExternalIDHandler,
		),
		BillingWebhookMiddleware: middleware.NewBillingWebhookMiddleware(env.StripeWebhookSecret),
		BillingWebhookHandler: httphandler.NewBillingWebhookHandler(
			checkoutCompletedHandler,
			subscriptionUpdatedHandler,
			subscriptionDeletedHandler,
			invoicePaymentSucceededHandler,
			invoicePaymentFailedHandler,
			upsertInvoiceHandler,
		),
		UserHandler: httphandler.NewUserHandler(getUserByIDHandler, setActiveProjectHandler),
		ProjectHandler: httphandler.NewProjectHandler(
			createOrgHandler,
			updateOrgHandler,
			deleteOrgHandler,
			removeOrgMemberHandler,
			getProjectByIDHandler,
			listProjectsByUserHandler,
			setActiveProjectHandler,
		),
		WorkflowHandler: httphandler.NewWorkflowHandler(
			createWorkflowHandler,
			updateWorkflowHandler,
			activateWorkflowHandler,
			deactivateWorkflowHandler,
			deleteWorkflowHandler,
			getWorkflowByIDHandler,
			listWorkflowsByProjectHandler,
			getProjectByIDHandler,
		),
		EndpointHandler: httphandler.NewEndpointHandler(
			createEndpointHandler,
			importEndpointsHandler,
			updateEndpointHandler,
			deleteEndpointHandler,
			getEndpointByIDHandler,
			listEndpointsByOrgHandler,
		),
		StepHandler: httphandler.NewStepHandler(
			createStepHandler,
			updateStepHandler,
			updateStepPositionHandler,
			deleteStepHandler,
			getStepByIDHandler,
			listStepsByWorkflowHandler,
			latestStepRunStatusesHandler,
			getWorkflowByIDHandler,
		),
		ConnectionHandler: httphandler.NewConnectionHandler(
			createConnHandler,
			deleteConnHandler,
			listConnsByWorkflowHandler,
			getWorkflowByIDHandler,
		),
		WorkflowRunHandler: httphandler.NewWorkflowRunHandler(
			startWorkflowRunHandler,
			cancelWorkflowRunHandler,
			getWorkflowRunByIDHandler,
			getWorkflowRunAnalyticsHandler,
			listWorkflowRunsHandler,
			listWorkflowRunsByProjectHandler,
			listStepRunsHandler,
			listStepRunsByIDsHandler,
			listInsightsByIDsHandler,
			getWorkflowByIDHandler,
			listStepsByWorkflowHandler,
		),
		VariableHandler: httphandler.NewVariableHandler(
			createVariableHandler,
			updateVariableHandler,
			deleteVariableHandler,
			getVariableByIDHandler,
			listVariablesByWorkflowHandler,
			listAvailableVariablesHandler,
			searchVariablePathsHandler,
			getStepByIDHandler,
			getWorkflowByIDHandler,
		),
		PlanHandler: httphandler.NewPlanHandler(listActivePlansHandler),
		SubscriptionHandler: httphandler.NewSubscriptionHandler(
			getCurrentSubscriptionHandler,
			getQuotaUsageHandler,
			previewPlanChangeHandler,
			createSubscriptionHandler,
			createBillingPortalHandler,
		),
		InvoiceHandler: httphandler.NewInvoiceHandler(listInvoicesHandler),
		RealtimeHandler: httphandler.NewRealtimeHandler(env),
	}
}
