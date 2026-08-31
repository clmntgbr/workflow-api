package quota_test

import (
	"testing"

	cmdquota "go-api/internal/application/command/quota"
)

func TestNormalizeExecutorPriority(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  uint8
	}{
		{name: "zero", input: 0, want: 0},
		{name: "negative", input: -1, want: 0},
		{name: "free plan", input: 0, want: 0},
		{name: "starter plan", input: 2, want: 2},
		{name: "pro plan", input: 5, want: 5},
		{name: "business plan", input: 10, want: 10},
		{name: "clamped above max", input: 99, want: cmdquota.MaxExecutorPriority},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := cmdquota.NormalizeExecutorPriority(tc.input); got != tc.want {
				t.Fatalf("priority: got %d want %d", got, tc.want)
			}
		})
	}
}
