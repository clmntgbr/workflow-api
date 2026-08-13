package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-api/cmd/scheduler/di"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	domainworkflowrun "go-api/internal/domain/workflowrun"
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

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	runScheduler(ctx, container, env.SchedulerInterval)
}

func runScheduler(ctx context.Context, container *di.Container, interval time.Duration) {
	if interval <= 0 {
		interval = time.Minute
	}

	log.Printf("scheduler started (interval=%s)", interval)
	tick(ctx, container)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("scheduler stopped")
			return
		case <-ticker.C:
			tick(ctx, container)
		}
	}
}

func tick(ctx context.Context, container *di.Container) {
	log.Println("scheduler: tick")
	workflows, err := container.WorkflowReadRepo.GetWorkflowsForExecution(ctx)
	if err != nil {
		log.Printf("scheduler: failed to list executable workflows: %v", err)
		return
	}

	log.Println("scheduler: workflows", len(workflows))
	for _, workflow := range workflows {
		run, err := container.StartWorkflowRunHandler.Handle(ctx, workflowruncmd.StartWorkflowRunCommand{
			WorkflowID:  workflow.ID,
			TriggeredBy: domainworkflowrun.TriggeredBySchedule,
		})
		if err != nil {
			if errors.Is(err, domainworkflowrun.ErrAlreadyInProgress) {
				continue
			}
			log.Printf("scheduler: failed to start workflow %s: %v", workflow.ID, err)
			continue
		}
		log.Printf("scheduler: started workflow run %s for workflow %s", run.ID, workflow.ID)
	}
	log.Println("scheduler: done")
}
