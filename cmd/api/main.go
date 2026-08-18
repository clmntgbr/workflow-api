package main

import (
	"go-api/cmd/api/di"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/schema"
	"log"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

func main() {
	env := config.Load()
	db := config.ConnectDatabase(env)

	if err := schema.AssertModelsMatchDB(db); err != nil {
		log.Fatalf("schema check failed: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:       "Go API",
		ServerHeader:  "Go API",
		CaseSensitive: true,
		StrictRouting: true,
		UnescapePath:  true,
		BodyLimit:     10 * 1024 * 1024,
	})

	app.Use(helmet.New())

	app.Use(cors.New(cors.Config{
		AllowOrigins:     env.CORSAllowedOrigins,
		AllowMethods:     env.CORSAllowMethods,
		AllowHeaders:     env.CORSAllowHeaders,
		AllowCredentials: env.CORSAllowCredentials,
		MaxAge:           env.CORSMaxAge,
	}))

	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Use(func(c fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		duration := time.Since(start)
		c.Append("Server-Timing", "app;dur="+duration.String())
		return err
	})

	container := di.NewContainer(db, env)
	setupRoutes(app, container)

	log.Println("Server is running on port", env.Port)
	log.Fatal(app.Listen(":" + env.Port))
}
