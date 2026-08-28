package main

import (
	"go-api/cmd/api/di"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/healthcheck"
)

func setupRoutes(app *fiber.App, container *di.Container) {
	setupHealthChecks(app)
	setupWebhooks(app, container)
	setupAPIRoutes(app, container)
}

func setupWebhooks(app *fiber.App, container *di.Container) {
	webhooks := app.Group("/webhooks")
	webhooks.Post("/clerk", container.UserWebhookMiddleware.Protected(), container.UserWebhookHandler.Execute)
	webhooks.Post(
		"/stripe",
		container.BillingWebhookMiddleware.Protected(),
		container.BillingWebhookHandler.Execute,
	)
}

func setupHealthChecks(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New())
	app.Get(healthcheck.StartupEndpoint, healthcheck.New())
}

func setupAPIRoutes(app *fiber.App, container *di.Container) {
	api := app.Group("/api")

	setupPlanRoutes(api, container)
	setupSubscriptionRoutes(api, container)
	setupInvoiceRoutes(api, container)
	setupUserRoutes(api, container)
	setupProjectRoutes(api, container)
	setupWorkflowRoutes(api, container)
	setupEndpointRoutes(api, container)
	setupStepRoutes(api, container)
	setupConnectionRoutes(api, container)
	setupVariableRoutes(api, container)
	setupAssertionRoutes(api, container)
	setupActivityLogRoutes(api, container)
	setupWorkflowRunRoutes(api, container)
	setupRealtimeRoutes(api, container)
}

func setupUserRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Get("/users/me", container.UserHandler.GetUser)
	api.Put("/users/me/active-project", container.UserHandler.SetActiveProject)
}

func setupProjectRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Get("/projects", container.ProjectHandler.List)
	api.Post("/projects", container.ProjectHandler.Create)
	api.Get("/projects/:id", container.ProjectHandler.GetByID)
	api.Put("/projects/:id", container.ProjectHandler.Update)
	api.Post("/projects/:id/activate", container.ProjectHandler.Activate)
	api.Delete("/projects/:id", container.ProjectHandler.Delete)
	api.Delete("/projects/:id/members/:userId", container.ProjectHandler.RemoveMember)
}

func setupWorkflowRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/workflows", container.WorkflowHandler.Create)
	api.Get("/workflows", container.WorkflowHandler.ListByProject)
	api.Get("/workflows/:id", container.WorkflowHandler.GetByID)
	api.Put("/workflows/:id", container.WorkflowHandler.Update)
	api.Post("/workflows/:id/activate", container.WorkflowHandler.Activate)
	api.Post("/workflows/:id/deactivate", container.WorkflowHandler.Deactivate)
	api.Delete("/workflows/:id", container.WorkflowHandler.Delete)
}

func setupEndpointRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/endpoints", container.EndpointHandler.Create)
	api.Post("/endpoints/import", container.EndpointHandler.ImportFromOpenAPI)
	api.Get("/endpoints", container.EndpointHandler.ListByProject)
	api.Get("/endpoints/:id", container.EndpointHandler.GetByID)
	api.Put("/endpoints/:id", container.EndpointHandler.Update)
	api.Delete("/endpoints/:id", container.EndpointHandler.Delete)
}

func setupStepRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/workflows/:workflowId/steps", container.StepHandler.Create)
	api.Get("/workflows/:workflowId/steps", container.StepHandler.ListByWorkflow)
	api.Get("/workflows/:workflowId/steps/:id", container.StepHandler.GetByID)
	api.Put("/workflows/:workflowId/steps/:id", container.StepHandler.Update)
	api.Put("/workflows/:workflowId/steps/:id/position", container.StepHandler.UpdatePosition)
	api.Delete("/workflows/:workflowId/steps/:id", container.StepHandler.Delete)
}

func setupConnectionRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/workflows/:workflowId/connections", container.ConnectionHandler.Create)
	api.Get("/workflows/:workflowId/connections", container.ConnectionHandler.ListByWorkflow)
	api.Delete("/workflows/:workflowId/connections/:id", container.ConnectionHandler.Delete)
}

func setupVariableRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/workflows/:workflowId/variables", container.VariableHandler.Create)
	api.Get("/workflows/:workflowId/variables", container.VariableHandler.List)
	api.Get("/workflows/:workflowId/steps/:stepId/variables", container.VariableHandler.ListAvailable)
	api.Get("/workflows/:workflowId/steps/:stepId/variable-paths", container.VariableHandler.SearchPaths)
	api.Get("/workflows/:workflowId/variables/:id", container.VariableHandler.GetByID)
	api.Put("/workflows/:workflowId/variables/:id", container.VariableHandler.Update)
	api.Delete("/workflows/:workflowId/variables/:id", container.VariableHandler.Delete)
}

func setupAssertionRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/workflows/:workflowId/steps/:stepId/assertions", container.AssertionHandler.Create)
	api.Get("/workflows/:workflowId/steps/:stepId/assertions", container.AssertionHandler.ListByStep)
	api.Get("/workflows/:workflowId/steps/:stepId/assertion-paths", container.AssertionHandler.SearchPaths)
	api.Get("/workflows/:workflowId/assertions/:id", container.AssertionHandler.GetByID)
	api.Put("/workflows/:workflowId/assertions/:id", container.AssertionHandler.Update)
	api.Delete("/workflows/:workflowId/assertions/:id", container.AssertionHandler.Delete)
}

func setupActivityLogRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Get("/workflows/:workflowId/activity", container.ActivityLogHandler.ListByWorkflow)
}

func setupWorkflowRunRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Get("/workflows/:workflowId/runs/analytics", container.WorkflowRunHandler.Analytics)
	api.Post("/workflows/:id/start", container.WorkflowRunHandler.StartWorkflow)
	api.Post("/workflows/:id/stop", container.WorkflowRunHandler.StopWorkflow)
	api.Get("/workflows/:workflowId/runs", container.WorkflowRunHandler.ListByWorkflow)
	api.Get("/workflows/:workflowId/runs/:id", container.WorkflowRunHandler.GetByID)
}

func setupPlanRoutes(api fiber.Router, container *di.Container) {
	api.Get("/plans", container.PlanHandler.List)
}

func setupSubscriptionRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/subscription", auth, container.SubscriptionHandler.GetSubscription)
	api.Get("/quota", auth, container.SubscriptionHandler.GetQuota)
	api.Post("/subscriptions", auth, container.SubscriptionHandler.CreateSubscription)
	api.Post("/subscriptions/preview", auth, container.SubscriptionHandler.PreviewSubscription)
	api.Get("/subscriptions/portal", auth, container.SubscriptionHandler.CreateBillingPortal)
}

func setupInvoiceRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/invoices", auth, container.InvoiceHandler.GetInvoices)
}

func setupRealtimeRoutes(api fiber.Router, container *di.Container) {
	auth := container.AuthenticateMiddleware.Protected()
	api.Get("/realtime/connection", auth, container.RealtimeHandler.GetConnection)
}
