package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"go-api/cmd/executor/di"
	"go-api/internal/infrastructure/config"
)

func main() {
	env := config.Load()
	container := di.NewContainer(env)
	defer container.Conn.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if err := container.Consumer.Start(ctx); err != nil {
		log.Fatalf("executor stopped with error: %v", err)
	}
}
