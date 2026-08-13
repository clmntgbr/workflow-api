package steprun

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	stepruncmd "go-api/internal/application/command/steprun"
	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"

	"github.com/google/uuid"
)

type ExecuteJob struct {
	StepRunID     string `json:"stepRunId"`
	StepID        string `json:"stepId"`
	WorkflowRunID string `json:"workflowRunId"`
}

type ExecuteHandler struct {
	stepRunRepo domainsteprun.StepRunWriteRepository
	http        port.HTTPExecutor
	start       *stepruncmd.StartStepRunHandler
	succeed     *stepruncmd.SucceedStepRunHandler
	fail        *stepruncmd.FailStepRunHandler
	increment   *stepruncmd.IncrementStepRunAttemptHandler
}

func NewExecuteHandler(
	stepRunRepo domainsteprun.StepRunWriteRepository,
	httpClient port.HTTPExecutor,
	start *stepruncmd.StartStepRunHandler,
	succeed *stepruncmd.SucceedStepRunHandler,
	fail *stepruncmd.FailStepRunHandler,
	increment *stepruncmd.IncrementStepRunAttemptHandler,
) *ExecuteHandler {
	return &ExecuteHandler{
		stepRunRepo: stepRunRepo,
		http:        httpClient,
		start:       start,
		succeed:     succeed,
		fail:        fail,
		increment:   increment,
	}
}

func (h *ExecuteHandler) Handle(ctx context.Context, payload []byte) error {
	var job ExecuteJob
	if err := json.Unmarshal(payload, &job); err != nil {
		return messaging.NonRetryable(err)
	}

	stepRunID, err := uuid.Parse(job.StepRunID)
	if err != nil {
		return messaging.NonRetryable(errInvalidExecuteJob)
	}
	if _, err := uuid.Parse(job.StepID); err != nil {
		return messaging.NonRetryable(errInvalidExecuteJob)
	}
	if _, err := uuid.Parse(job.WorkflowRunID); err != nil {
		return messaging.NonRetryable(errInvalidExecuteJob)
	}

	run, err := h.stepRunRepo.GetByID(ctx, stepRunID)
	if err != nil {
		return messaging.Retryable(err)
	}
	if run == nil {
		return messaging.NonRetryable(domainsteprun.ErrNotFound)
	}
	if run.Status.IsTerminal() {
		return nil
	}

	run, err = h.start.Handle(ctx, stepruncmd.StartStepRunCommand{StepRunID: stepRunID})
	if err != nil {
		if errors.Is(err, domainsteprun.ErrNotFound) {
			return messaging.NonRetryable(err)
		}
		return messaging.Retryable(err)
	}
	if run.Status.IsTerminal() {
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			return messaging.Retryable(ctx.Err())
		default:
		}

		response, execErr := h.http.Do(ctx, port.HTTPRequest{
			Method:  run.Method,
			URL:     run.URL,
			Headers: run.Headers,
			Query:   run.Query,
			Body:    run.Body,
			Timeout: timeoutDuration(run.Timeout),
		})
		if execErr == nil {
			snapshot := domainsteprun.ResponseSnapshot{
				Status:  response.Status,
				Headers: response.Headers,
				Body:    response.Body,
			}
			_, err = h.succeed.Handle(ctx, stepruncmd.SucceedStepRunCommand{
				StepRunID: stepRunID,
				Response:  snapshot,
			})
			if err != nil {
				return messaging.Retryable(err)
			}
			log.Printf(
				"executor succeeded stepRunId=%s attempt=%d status=%d",
				stepRunID,
				run.Attempt,
				response.Status,
			)
			return nil
		}

		errMsg := execErr.Error()
		var snapshot *domainsteprun.ResponseSnapshot
		if response != nil {
			s := domainsteprun.ResponseSnapshot{
				Status:  response.Status,
				Headers: response.Headers,
				Body:    response.Body,
			}
			snapshot = &s
		}

		if !run.CanRetry() {
			_, err = h.fail.Handle(ctx, stepruncmd.FailStepRunCommand{
				StepRunID: stepRunID,
				Error:     errMsg,
				Response:  snapshot,
			})
			if err != nil {
				return messaging.Retryable(err)
			}
			log.Printf(
				"executor failed stepRunId=%s attempt=%d err=%s",
				stepRunID,
				run.Attempt,
				errMsg,
			)
			return nil
		}

		run, err = h.increment.Handle(ctx, stepruncmd.IncrementStepRunAttemptCommand{
			StepRunID: stepRunID,
			Response:  snapshot,
			Error:     errMsg,
		})
		if err != nil {
			return messaging.Retryable(err)
		}

		log.Printf(
			"executor retrying stepRunId=%s nextAttempt=%d after=%dms err=%s",
			stepRunID,
			run.Attempt,
			run.RetryDelay,
			errMsg,
		)

		if err := sleep(ctx, retryDelayDuration(run.RetryDelay)); err != nil {
			return messaging.Retryable(err)
		}
	}
}

func timeoutDuration(timeoutMS int) time.Duration {
	if timeoutMS <= 0 {
		return 30 * time.Second
	}
	return time.Duration(timeoutMS) * time.Millisecond
}

func retryDelayDuration(delayMS int) time.Duration {
	if delayMS < 0 {
		delayMS = 0
	}
	return time.Duration(delayMS) * time.Millisecond
}

func sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type invalidExecuteJobError struct{}

func (invalidExecuteJobError) Error() string { return "invalid step run execute job" }

var errInvalidExecuteJob error = invalidExecuteJobError{}
