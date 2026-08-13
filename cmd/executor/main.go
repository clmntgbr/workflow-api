package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-api/cmd/executor/di"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/schema"
)

func main() {
	env := config.Load()
	db := config.ConnectDatabase(env)

	if err := schema.AssertModelsMatchDB(db); err != nil {
		log.Fatalf("schema check failed: %v", err)
	}

	container := di.NewContainer(db, env)
	defer container.Conn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := container.Consumer.Start(ctx); err != nil {
		log.Fatalf("executor stopped with error: %v", err)
	}
}
