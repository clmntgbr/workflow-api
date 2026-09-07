package steprun

import (
	"testing"
	"time"

	domainstep "go-api/internal/domain/step"
)

func TestElapsedMS(t *testing.T) {
	start := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	end := start.Add(1500 * time.Millisecond)

	if got := ElapsedMS(&start, &end); got != 1500 {
		t.Fatalf("got %d want 1500", got)
	}
	if got := ElapsedMS(nil, &end); got != 0 {
		t.Fatalf("nil start: got %d", got)
	}
	if got := ElapsedMS(&start, nil); got != 0 {
		t.Fatalf("nil end: got %d", got)
	}
	if got := ElapsedMS(&end, &start); got != 0 {
		t.Fatalf("inverted range: got %d", got)
	}
}

func TestSumExecutionElapsedMS_SumsHTTPDelayAndCondition(t *testing.T) {
	start := time.Date(2026, 1, 2, 10, 0, 0, 0, time.UTC)
	httpEnd := start.Add(2 * time.Second)
	delayEnd := start.Add(time.Hour)
	conditionEnd := start.Add(10 * time.Millisecond)

	total := SumExecutionElapsedMS([]StepRunView{
		{
			StepType:   domainstep.TypeHTTP,
			StartedAt:  &start,
			FinishedAt: &httpEnd,
		},
		{
			StepType:   domainstep.TypeDelay,
			StartedAt:  &start,
			FinishedAt: &delayEnd,
		},
		{
			StepType:   domainstep.TypeCondition,
			StartedAt:  &start,
			FinishedAt: &conditionEnd,
		},
		{
			StepType:  domainstep.TypeHTTP,
			StartedAt: &start,
		},
	})
	want := int64(2_000 + 3_600_000 + 10)
	if total != want {
		t.Fatalf("got %d want %d (HTTP + delay + condition; unfinished HTTP ignored)", total, want)
	}
}
