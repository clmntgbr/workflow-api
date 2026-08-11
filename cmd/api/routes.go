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
}

func setupHealthChecks(app *fiber.App) {
	app.Get(healthcheck.LivenessEndpoint, healthcheck.New())
	app.Get(healthcheck.ReadinessEndpoint, healthcheck.New())
	app.Get(healthcheck.StartupEndpoint, healthcheck.New())
}

func setupAPIRoutes(app *fiber.App, container *di.Container) {
	api := app.Group("/api")

	setupUserRoutes(api, container)
	setupOrganizationRoutes(api, container)
	setupWorkflowRoutes(api, container)
}

func setupUserRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Get("/users/me", container.UserHandler.GetUser)
}

func setupOrganizationRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/organizations", container.OrganizationHandler.Create)
	api.Get("/organizations/:id", container.OrganizationHandler.GetByID)
	api.Put("/organizations/:id", container.OrganizationHandler.Update)
	api.Delete("/organizations/:id", container.OrganizationHandler.Delete)
	api.Post("/organizations/:id/members", container.OrganizationHandler.AddMember)
	api.Delete("/organizations/:id/members/:userId", container.OrganizationHandler.RemoveMember)
}

func setupWorkflowRoutes(api fiber.Router, container *di.Container) {
	api.Use(container.AuthenticateMiddleware.Protected())
	api.Post("/workflows", container.WorkflowHandler.Create)
	api.Get("/workflows", container.WorkflowHandler.ListByOrganization)
	api.Get("/workflows/:id", container.WorkflowHandler.GetByID)
	api.Put("/workflows/:id", container.WorkflowHandler.Update)
	api.Delete("/workflows/:id", container.WorkflowHandler.Delete)
}
