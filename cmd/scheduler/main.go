package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go-api/cmd/scheduler/di"
	cmdquota "go-api/internal/application/command/quota"
	workflowruncmd "go-api/internal/application/command/workflowrun"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"
	"go-api/internal/infrastructure/config"
	"go-api/internal/infrastructure/persistence/schema"

	"github.com/google/uuid"
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

	log.Printf(
		"scheduler started (interval=%s batchSize=%d concurrency=%d maxBatches=%d)",
		interval,
		container.BatchSize,
		container.Concurrency,
		container.MaxBatchesPerTick,
	)
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
	startedAt := time.Now().UTC()
	var started, skipped, failed, claimedTotal int64

	for batch := 0; batch < container.MaxBatchesPerTick; batch++ {
		if ctx.Err() != nil {
			break
		}

		claimed, err := container.WorkflowWriteRepo.ClaimDueForExecution(
			ctx,
			time.Now().UTC(),
			container.BatchSize,
		)
		if err != nil {
			log.Printf("scheduler: claim failed: %v", err)
			break
		}
		if len(claimed) == 0 {
			break
		}
		claimedTotal += int64(len(claimed))

		processClaimedBatch(ctx, container, claimed, &started, &skipped, &failed)

		if len(claimed) < container.BatchSize {
			break
		}
	}

	log.Printf(
		"scheduler: tick done claimed=%d started=%d skipped=%d failed=%d duration=%s",
		claimedTotal,
		started,
		skipped,
		failed,
		time.Since(startedAt).Round(time.Millisecond),
	)
}

func processClaimedBatch(
	ctx context.Context,
	container *di.Container,
	claimed []*domainworkflow.Workflow,
	started, skipped, failed *int64,
) {
	sem := make(chan struct{}, container.Concurrency)
	var wg sync.WaitGroup

	for _, workflow := range claimed {
		wfID := workflow.ID
		wg.Add(1)
		sem <- struct{}{}
		go func(workflowID uuid.UUID) {
			defer wg.Done()
			defer func() { <-sem }()

			_, err := container.StartWorkflowRunHandler.Handle(ctx, workflowruncmd.StartWorkflowRunCommand{
				WorkflowID:              workflowID,
				TriggeredBy:             domainworkflowrun.TriggeredBySchedule,
				ScheduleAlreadyAdvanced: true,
			})
			if err != nil {
				if errors.Is(err, domainworkflowrun.ErrAlreadyInProgress) ||
					errors.Is(err, cmdquota.ErrWorkflowRunQuotaExceeded) ||
					errors.Is(err, cmdquota.ErrConcurrentRunQuotaExceeded) {
					atomic.AddInt64(skipped, 1)
					return
				}
				atomic.AddInt64(failed, 1)
				log.Printf("scheduler: failed to start workflow %s: %v", workflowID, err)
				return
			}
			atomic.AddInt64(started, 1)
		}(wfID)
	}

	wg.Wait()
}
