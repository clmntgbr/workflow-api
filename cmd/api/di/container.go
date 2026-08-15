package di

import (
	"log"

	authcmd "go-api/internal/application/command/auth"
	conncmd "go-api/internal/application/command/connection"
	endpointcmd "go-api/internal/application/command/endpoint"
	identitycmd "go-api/internal/application/command/identity"
	orgcmd "go-api/internal/application/command/organization"
	stepcmd "go-api/internal/application/command/step"
	usercmd "go-api/internal/application/command/user"
	variablecmd "go-api/internal/application/command/variable"
	workflowcmd "go-api/internal/application/command/workflow"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	queryconn "go-api/internal/application/query/connection"
	queryendpoint "go-api/internal/application/query/endpoint"
	queryinsight "go-api/internal/application/query/insight"
	queryorganization "go-api/internal/application/query/organization"
	querystep "go-api/internal/application/query/step"
	querysteprun "go-api/internal/application/query/steprun"
	queryuser "go-api/internal/application/query/user"
	queryvariable "go-api/internal/application/query/variable"
	queryworkflow "go-api/internal/application/query/workflow"
	queryworkflowrun "go-api/internal/application/query/workflowrun"
	infraClerk "go-api/internal/infrastructure/clerk"
	"go-api/internal/infrastructure/config"
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
	UserHandler            *httphandler.UserHandler
	OrganizationHandler    *httphandler.OrganizationHandler
	WorkflowHandler        *httphandler.WorkflowHandler
	EndpointHandler        *httphandler.EndpointHandler
	StepHandler            *httphandler.StepHandler
	ConnectionHandler      *httphandler.ConnectionHandler
	WorkflowRunHandler     *httphandler.WorkflowRunHandler
	VariableHandler        *httphandler.VariableHandler
	RealtimeHandler        *httphandler.RealtimeHandler
}

func NewContainer(db *gorm.DB, env *config.Config) *Container {
	jwksProvider, err := infraClerk.NewJWKSProvider(env)
	if err != nil {
		log.Fatalf("failed to create JWKS provider: %v", err)
	}

	userWriteRepo := write.NewUserWriteRepository(db)
	userReadRepo := read.NewUserReadRepository(db)
	orgWriteRepo := write.NewOrganizationWriteRepository(db)
	orgReadRepo := read.NewOrganizationReadRepository(db)
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
	stepRunReadRepo := read.NewStepRunReadRepository(db)
	insightReadRepo := read.NewInsightReadRepository(db)
	variableWriteRepo := write.NewVariableWriteRepository(db)
	variableReadRepo := read.NewVariableReadRepository(db)
	outboxRepo := outbox.NewRepository(db)

	createUserHandler := usercmd.NewCreateUserHandler(userWriteRepo, orgWriteRepo, outboxRepo)
	updateUserHandler := usercmd.NewUpdateUserHandler(userWriteRepo, outboxRepo)
	getUserByExternalIDHandler := usercmd.NewGetUserByExternalIDHandler(userWriteRepo)
	deleteUserByExternalIDHandler := usercmd.NewDeleteUserByExternalIDHandler(userWriteRepo, outboxRepo)
	setActiveOrganizationHandler := usercmd.NewSetActiveOrganizationHandler(userWriteRepo, orgWriteRepo, outboxRepo)
	validateTokenHandler := authcmd.NewValidateTokenHandler(jwksProvider, userWriteRepo)
	fetchUserHandler := identitycmd.NewFetchUserHandler(infraClerk.NewUserGateway(env.ClerkSecretKey))
	getUserByIDHandler := queryuser.NewGetUserByIDHandler(userReadRepo)

	createOrgHandler := orgcmd.NewCreateOrganizationHandler(orgWriteRepo, userWriteRepo, outboxRepo)
	updateOrgHandler := orgcmd.NewUpdateOrganizationHandler(orgWriteRepo, outboxRepo)
	deleteOrgHandler := orgcmd.NewDeleteOrganizationHandler(orgWriteRepo, outboxRepo)
	addOrgMemberHandler := orgcmd.NewAddOrganizationMemberHandler(orgWriteRepo, userWriteRepo, outboxRepo)
	removeOrgMemberHandler := orgcmd.NewRemoveOrganizationMemberHandler(orgWriteRepo, userWriteRepo, outboxRepo)
	getOrgByIDHandler := queryorganization.NewGetOrganizationByIDHandler(orgReadRepo)
	listOrgsByUserHandler := queryorganization.NewListOrganizationsByUserHandler(orgReadRepo)

	createWorkflowHandler := workflowcmd.NewCreateWorkflowHandler(workflowWriteRepo, outboxRepo)
	updateWorkflowHandler := workflowcmd.NewUpdateWorkflowHandler(workflowWriteRepo, outboxRepo)
	deleteWorkflowHandler := workflowcmd.NewDeleteWorkflowHandler(workflowWriteRepo, outboxRepo)
	getWorkflowByIDHandler := queryworkflow.NewGetWorkflowByIDHandler(workflowReadRepo)
	listWorkflowsByOrgHandler := queryworkflow.NewListWorkflowsByOrganizationHandler(workflowReadRepo)

	createEndpointHandler := endpointcmd.NewCreateEndpointHandler(endpointWriteRepo, outboxRepo)
	updateEndpointHandler := endpointcmd.NewUpdateEndpointHandler(endpointWriteRepo, outboxRepo)
	deleteEndpointHandler := endpointcmd.NewDeleteEndpointHandler(endpointWriteRepo, outboxRepo)
	getEndpointByIDHandler := queryendpoint.NewGetEndpointByIDHandler(endpointReadRepo)
	listEndpointsByOrgHandler := queryendpoint.NewListEndpointsByOrganizationHandler(endpointReadRepo)

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

	createStepHandler := stepcmd.NewCreateStepHandler(
		stepWriteRepo,
		stepReadRepo,
		connReadRepo,
		endpointReadRepo,
		workflowReadRepo,
		outboxRepo,
	)
	updateStepHandler := stepcmd.NewUpdateStepHandler(stepWriteRepo, connReadRepo, variableReadRepo, outboxRepo)
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
	deleteVariableHandler := variablecmd.NewDeleteVariableHandler(variableWriteRepo)
	getVariableByIDHandler := queryvariable.NewGetVariableByIDHandler(variableReadRepo)
	listVariablesByWorkflowHandler := queryvariable.NewListVariablesByWorkflowHandler(variableReadRepo)
	listAvailableVariablesHandler := queryvariable.NewListAvailableVariablesHandler(variableReadRepo, connReadRepo)
	searchVariablePathsHandler := queryvariable.NewSearchVariablePathsHandler(stepRunReadRepo)

	startWorkflowRunHandler := workflowruncmd.NewStartWorkflowRunHandler(
		workflowWriteRepo,
		workflowRunWriteRepo,
		outboxRepo,
	)
	getWorkflowRunByIDHandler := queryworkflowrun.NewGetWorkflowRunByIDHandler(workflowRunReadRepo)
	listWorkflowRunsHandler := queryworkflowrun.NewListWorkflowRunsByWorkflowHandler(workflowRunReadRepo)
	listWorkflowRunsByOrgHandler := queryworkflowrun.NewListWorkflowRunsByOrganizationHandler(workflowRunReadRepo)
	listStepRunsHandler := querysteprun.NewListStepRunsByWorkflowRunHandler(stepRunReadRepo)
	listStepRunsByIDsHandler := querysteprun.NewListStepRunsByWorkflowRunIDsHandler(stepRunReadRepo)
	listInsightsByIDsHandler := queryinsight.NewListInsightsByStepRunIDsHandler(insightReadRepo)

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
		UserHandler: httphandler.NewUserHandler(getUserByIDHandler, setActiveOrganizationHandler),
		OrganizationHandler: httphandler.NewOrganizationHandler(
			createOrgHandler,
			updateOrgHandler,
			deleteOrgHandler,
			addOrgMemberHandler,
			removeOrgMemberHandler,
			getOrgByIDHandler,
			listOrgsByUserHandler,
			setActiveOrganizationHandler,
		),
		WorkflowHandler: httphandler.NewWorkflowHandler(
			createWorkflowHandler,
			updateWorkflowHandler,
			deleteWorkflowHandler,
			getWorkflowByIDHandler,
			listWorkflowsByOrgHandler,
			getOrgByIDHandler,
		),
		EndpointHandler: httphandler.NewEndpointHandler(
			createEndpointHandler,
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
			getWorkflowRunByIDHandler,
			listWorkflowRunsHandler,
			listWorkflowRunsByOrgHandler,
			listStepRunsHandler,
			listStepRunsByIDsHandler,
			listInsightsByIDsHandler,
			getWorkflowByIDHandler,
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
		RealtimeHandler: httphandler.NewRealtimeHandler(env),
	}
}
