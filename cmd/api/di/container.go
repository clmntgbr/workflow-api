package di

import (
	"log"

	authcmd "go-api/internal/application/command/auth"
	identitycmd "go-api/internal/application/command/identity"
	orgcmd "go-api/internal/application/command/organization"
	usercmd "go-api/internal/application/command/user"
	workflowcmd "go-api/internal/application/command/workflow"
	queryorganization "go-api/internal/application/query/organization"
	queryuser "go-api/internal/application/query/user"
	queryworkflow "go-api/internal/application/query/workflow"
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
	outboxRepo := outbox.NewRepository(db)

	createUserHandler := usercmd.NewCreateUserHandler(userWriteRepo, outboxRepo)
	updateUserHandler := usercmd.NewUpdateUserHandler(userWriteRepo, outboxRepo)
	getUserByExternalIDHandler := usercmd.NewGetUserByExternalIDHandler(userWriteRepo)
	deleteUserByExternalIDHandler := usercmd.NewDeleteUserByExternalIDHandler(userWriteRepo, outboxRepo)
	validateTokenHandler := authcmd.NewValidateTokenHandler(jwksProvider, userWriteRepo)
	fetchUserHandler := identitycmd.NewFetchUserHandler(infraClerk.NewUserGateway(env.ClerkSecretKey))
	getUserByIDHandler := queryuser.NewGetUserByIDHandler(userReadRepo)

	createOrgHandler := orgcmd.NewCreateOrganizationHandler(orgWriteRepo, outboxRepo)
	updateOrgHandler := orgcmd.NewUpdateOrganizationHandler(orgWriteRepo, outboxRepo)
	deleteOrgHandler := orgcmd.NewDeleteOrganizationHandler(orgWriteRepo, outboxRepo)
	addOrgMemberHandler := orgcmd.NewAddOrganizationMemberHandler(orgWriteRepo, outboxRepo)
	removeOrgMemberHandler := orgcmd.NewRemoveOrganizationMemberHandler(orgWriteRepo, outboxRepo)
	getOrgByIDHandler := queryorganization.NewGetOrganizationByIDHandler(orgReadRepo)

	createWorkflowHandler := workflowcmd.NewCreateWorkflowHandler(workflowWriteRepo, outboxRepo)
	updateWorkflowHandler := workflowcmd.NewUpdateWorkflowHandler(workflowWriteRepo, outboxRepo)
	deleteWorkflowHandler := workflowcmd.NewDeleteWorkflowHandler(workflowWriteRepo, outboxRepo)
	getWorkflowByIDHandler := queryworkflow.NewGetWorkflowByIDHandler(workflowReadRepo)
	listWorkflowsByOrgHandler := queryworkflow.NewListWorkflowsByOrganizationHandler(workflowReadRepo)

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
		UserHandler: httphandler.NewUserHandler(getUserByIDHandler),
		OrganizationHandler: httphandler.NewOrganizationHandler(
			createOrgHandler,
			updateOrgHandler,
			deleteOrgHandler,
			addOrgMemberHandler,
			removeOrgMemberHandler,
			getOrgByIDHandler,
		),
		WorkflowHandler: httphandler.NewWorkflowHandler(
			createWorkflowHandler,
			updateWorkflowHandler,
			deleteWorkflowHandler,
			getWorkflowByIDHandler,
			listWorkflowsByOrgHandler,
		),
	}
}
