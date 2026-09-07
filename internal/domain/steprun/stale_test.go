package steprun

import (
	"testing"
	"time"

	domainstep "go-api/internal/domain/step"
)

func TestStepRun_IsStale_IgnoresDelayAndWaiting(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	old := now.Add(-2 * time.Hour)

	delay := &StepRun{StepType: domainstep.TypeDelay, Status: StatusPending, CreatedAt: old}
	if delay.IsStale(now, 30*time.Minute, time.Minute) {
		t.Fatal("delay pending must not be stale")
	}

	waiting := &StepRun{StepType: domainstep.TypeHTTP, Status: StatusWaiting, CreatedAt: old}
	if waiting.IsStale(now, 30*time.Minute, time.Minute) {
		t.Fatal("waiting must not be stale")
	}

	done := &StepRun{StepType: domainstep.TypeHTTP, Status: StatusSuccess, CreatedAt: old}
	if done.IsStale(now, 30*time.Minute, time.Minute) {
		t.Fatal("terminal must not be stale")
	}
}

func TestStepRun_IsStale_PendingUsesQueueAge(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	pendingMaxAge := 30 * time.Minute

	fresh := &StepRun{
		StepType:  domainstep.TypeHTTP,
		Status:    StatusPending,
		CreatedAt: now.Add(-10 * time.Minute),
		Timeout:   7200000,
	}
	if fresh.IsStale(now, pendingMaxAge, time.Minute) {
		t.Fatal("pending HTTP younger than pendingMaxAge must not be stale")
	}

	old := &StepRun{
		StepType:  domainstep.TypeHTTP,
		Status:    StatusPending,
		CreatedAt: now.Add(-31 * time.Minute),
		Timeout:   7200000,
	}
	if !old.IsStale(now, pendingMaxAge, time.Minute) {
		t.Fatal("pending HTTP older than pendingMaxAge must be stale even with a long timeout")
	}
}

func TestStepRun_IsStale_RunningHonorsTimeoutAndRetries(t *testing.T) {
	now := time.Date(2026, 9, 7, 12, 0, 0, 0, time.UTC)
	started := now.Add(-40 * time.Minute)
	grace := 2 * time.Minute

	withinBudget := &StepRun{
		StepType:       domainstep.TypeHTTP,
		Status:         StatusRunning,
		Timeout:        3600000,
		RetryOnFailure: false,
		StartedAt:      &started,
		CreatedAt:      started,
	}
	if withinBudget.IsStale(now, 30*time.Minute, grace) {
		t.Fatal("running HTTP still inside timeout must not be stale")
	}

	retryStarted := now.Add(-9 * time.Minute)
	retriesOK := &StepRun{
		StepType:       domainstep.TypeHTTP,
		Status:         StatusRunning,
		Timeout:        120000,
		RetryOnFailure: true,
		RetryCount:     3,
		RetryDelay:     60000,
		StartedAt:      &retryStarted,
		CreatedAt:      retryStarted,
	}
	if retriesOK.IsStale(now, 30*time.Minute, grace) {
		t.Fatal("running HTTP inside timeout*attempts + retry delays must not be stale")
	}

	expiredStarted := now.Add(-15 * time.Minute)
	expired := &StepRun{
		StepType:       domainstep.TypeHTTP,
		Status:         StatusRunning,
		Timeout:        30000,
		RetryOnFailure: false,
		StartedAt:      &expiredStarted,
		CreatedAt:      expiredStarted,
	}
	if !expired.IsStale(now, 30*time.Minute, grace) {
		t.Fatal("running HTTP past timeout + grace must be stale")
	}
}

func TestStepRun_ExecutionBudget(t *testing.T) {
	noRetry := &StepRun{Timeout: 30000}
	if noRetry.ExecutionBudget() != 30*time.Second {
		t.Fatalf("got %s", noRetry.ExecutionBudget())
	}

	retries := &StepRun{Timeout: 10000, RetryOnFailure: true, RetryCount: 3, RetryDelay: 2000}
	want := 10*time.Second*3 + 2*time.Second*2
	if retries.ExecutionBudget() != want {
		t.Fatalf("got %s want %s", retries.ExecutionBudget(), want)
	}

	zeroTimeout := &StepRun{Timeout: 0}
	if zeroTimeout.ExecutionBudget() != 30*time.Second {
		t.Fatalf("zero timeout should fall back to default, got %s", zeroTimeout.ExecutionBudget())
	}
}
