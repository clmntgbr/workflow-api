package subscription_test

import (
	"testing"
	"time"

	querysubscription "go-api/internal/application/query/subscription"
)

func TestRunHistoryCutoff(t *testing.T) {
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)

	if got := querysubscription.RunHistoryCutoff(0, now); got != nil {
		t.Fatalf("expected nil cutoff for zero retention, got %v", got)
	}
	if got := querysubscription.RunHistoryCutoff(-1, now); got != nil {
		t.Fatalf("expected nil cutoff for negative retention, got %v", got)
	}

	cutoff := querysubscription.RunHistoryCutoff(7, now)
	if cutoff == nil {
		t.Fatal("expected cutoff")
	}
	want := time.Date(2026, 8, 24, 12, 0, 0, 0, time.UTC)
	if !cutoff.Equal(want) {
		t.Fatalf("cutoff: got %s want %s", cutoff, want)
	}
}

func TestClampTimeFrom(t *testing.T) {
	cutoff := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	after := time.Date(2026, 8, 15, 0, 0, 0, 0, time.UTC)

	if got := querysubscription.ClampTimeFrom(nil, nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
	if got := querysubscription.ClampTimeFrom(&after, nil); got != &after {
		t.Fatalf("expected requested time unchanged")
	}
	if got := querysubscription.ClampTimeFrom(nil, &cutoff); got != &cutoff {
		t.Fatalf("expected cutoff when from is nil")
	}

	before := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if got := querysubscription.ClampTimeFrom(&before, &cutoff); got != &cutoff {
		t.Fatalf("expected cutoff when from is before retention window")
	}
	if got := querysubscription.ClampTimeFrom(&after, &cutoff); got != &after {
		t.Fatalf("expected requested from when after cutoff")
	}
}

func TestIsWithinRunHistory(t *testing.T) {
	cutoff := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)

	if !querysubscription.IsWithinRunHistory(time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC), nil) {
		t.Fatal("expected within history when cutoff is nil")
	}
	if querysubscription.IsWithinRunHistory(time.Date(2026, 7, 31, 23, 59, 59, 0, time.UTC), &cutoff) {
		t.Fatal("expected outside history before cutoff")
	}
	if !querysubscription.IsWithinRunHistory(cutoff, &cutoff) {
		t.Fatal("expected within history at cutoff boundary")
	}
}
