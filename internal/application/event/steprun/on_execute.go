package steprun

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	insightcmd "go-api/internal/application/command/insight"
	stepruncmd "go-api/internal/application/command/steprun"
	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type ExecuteJob struct {
	StepRunID     string `json:"stepRunId"`
	StepID        string `json:"stepId"`
	WorkflowRunID string `json:"workflowRunId"`
}

type ExecuteHandler struct {
	stepRunRepo   domainsteprun.StepRunWriteRepository
	http          port.HTTPExecutor
	start         *stepruncmd.StartStepRunHandler
	succeed       *stepruncmd.SucceedStepRunHandler
	fail          *stepruncmd.FailStepRunHandler
	increment     *stepruncmd.IncrementStepRunAttemptHandler
	createInsight *insightcmd.CreateInsightHandler
}

func NewExecuteHandler(
	stepRunRepo domainsteprun.StepRunWriteRepository,
	httpClient port.HTTPExecutor,
	start *stepruncmd.StartStepRunHandler,
	succeed *stepruncmd.SucceedStepRunHandler,
	fail *stepruncmd.FailStepRunHandler,
	increment *stepruncmd.IncrementStepRunAttemptHandler,
	createInsight *insightcmd.CreateInsightHandler,
) *ExecuteHandler {
	return &ExecuteHandler{
		stepRunRepo:   stepRunRepo,
		http:          httpClient,
		start:         start,
		succeed:       succeed,
		fail:          fail,
		increment:     increment,
		createInsight: createInsight,
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

	attemptQueueAnchor := run.CreatedAt
	if run.StartedAt != nil {
		attemptQueueAnchor = *run.StartedAt
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

		timing := port.HTTPTiming{}
		if response != nil {
			timing = response.Timing
		}
		if timing.StartTime.IsZero() {
			timing.StartTime = time.Now().UTC()
		}
		if timing.EndTime.IsZero() {
			timing.EndTime = time.Now().UTC()
		}

		queueTime := timing.StartTime.Sub(attemptQueueAnchor)
		if queueTime < 0 {
			queueTime = 0
		}

		errMsg := ""
		errType := ""
		hasStatus := false
		statusCode := 0
		var snapshot *domainsteprun.ResponseSnapshot

		if execErr == nil {
			hasStatus = true
			statusCode = response.Status
			s := domainsteprun.ResponseSnapshot{
				Status:  response.Status,
				Headers: response.Headers,
				Body:    response.Body,
			}
			snapshot = &s

			// Transport OK is not enough: non-2xx responses fail the step run.
			if response.Status < 200 || response.Status >= 300 {
				execErr = fmt.Errorf("HTTP %d", response.Status)
				errMsg = execErr.Error()
				errType = "http_status"
			}
		} else {
			errMsg = execErr.Error()
			errType = timing.ErrorType
			if errType == "" {
				errType = "unknown"
			}
			if response != nil && response.Status > 0 {
				hasStatus = true
				statusCode = response.Status
				s := domainsteprun.ResponseSnapshot{
					Status:  response.Status,
					Headers: response.Headers,
					Body:    response.Body,
				}
				snapshot = &s
			}
		}

		if err := h.saveInsight(ctx, run, timing, queueTime, statusCode, hasStatus, errMsg, errType); err != nil {
			return messaging.Retryable(err)
		}

		if execErr == nil {
			extracted := extractVariables(run.VariableExtracts, response.Body)
			_, err = h.succeed.Handle(ctx, stepruncmd.SucceedStepRunCommand{
				StepRunID:          stepRunID,
				Response:           *snapshot,
				ExtractedVariables: extracted,
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

		delay := retryDelayDuration(run.RetryDelay)
		attemptQueueAnchor = time.Now().UTC()
		if err := sleep(ctx, delay); err != nil {
			return messaging.Retryable(err)
		}
	}
}

func (h *ExecuteHandler) saveInsight(
	ctx context.Context,
	run *domainsteprun.StepRun,
	timing port.HTTPTiming,
	queueTime time.Duration,
	statusCode int,
	hasStatus bool,
	errMsg string,
	errType string,
) error {
	totalAttempts := 1
	if run.RetryOnFailure && run.RetryCount > 0 {
		totalAttempts = run.RetryCount
	}

	_, err := h.createInsight.Handle(ctx, insightcmd.CreateInsightCommand{
		StepRunID:         run.ID,
		StartTime:         timing.StartTime,
		EndTime:           timing.EndTime,
		QueueTime:         queueTime,
		DNSLookupDuration: timing.DNSLookupDuration,
		TCPConnectionTime: timing.TCPConnectionTime,
		TLSHandshakeTime:  timing.TLSHandshakeTime,
		TTFB:              timing.TTFB,
		Duration:          timing.Duration,
		StatusCode:        statusCode,
		HasStatusCode:     hasStatus,
		ResponseSize:      timing.ResponseSize,
		RequestSize:       timing.RequestSize,
		AttemptNumber:     run.Attempt,
		TotalAttempts:     totalAttempts,
		ErrorMessage:      errMsg,
		ErrorType:         errType,
	})
	return err
}

func extractVariables(extracts []domainsteprun.VariableExtract, body any) map[string]any {
	out := map[string]any{}
	for _, extract := range extracts {
		value, err := domainvariable.ExtractByPath(body, extract.Path)
		if err != nil {
			continue
		}
		out[extract.VariableID.String()] = value
	}
	return out
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
