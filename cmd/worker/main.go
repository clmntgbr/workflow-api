package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api/cmd/worker/di"
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

	go container.Relay.Start(ctx)
	go runWaitingStepRunPoller(ctx, container)

	if err := container.Consumer.Start(ctx); err != nil {
		log.Fatalf("worker stopped with error: %v", err)
	}
}

func runWaitingStepRunPoller(ctx context.Context, container *di.Container) {
	log.Printf(
		"worker: waiting step run poller started (interval=%s batchSize=%d)",
		container.WaitingPollInterval,
		container.WaitingPollBatchSize,
	)

	poll := func() {
		resumed, err := container.ResumeWaitingStepRunsHandler.Handle(
			ctx,
			time.Now().UTC(),
			container.WaitingPollBatchSize,
		)
		if err != nil {
			log.Printf("worker: resume waiting step runs failed: %v", err)
			return
		}
		if resumed > 0 {
			log.Printf("worker: resumed %d waiting step run(s)", resumed)
		}
	}

	poll()

	ticker := time.NewTicker(container.WaitingPollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("worker: waiting step run poller stopped")
			return
		case <-ticker.C:
			poll()
		}
	}
}
