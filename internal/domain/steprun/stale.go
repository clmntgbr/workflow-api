package steprun

import (
	"time"

	domainstep "go-api/internal/domain/step"
)

const DefaultHTTPTimeoutMS = 30_000

func (s *StepRun) IsStale(now time.Time, pendingMaxAge, grace time.Duration) bool {
	if s == nil {
		return false
	}
	if s.StepType == domainstep.TypeDelay {
		return false
	}
	if s.Status.IsTerminal() || s.Status == StatusWaiting {
		return false
	}

	now = now.UTC()
	switch s.Status {
	case StatusPending:
		if pendingMaxAge <= 0 {
			return false
		}
		return !s.CreatedAt.UTC().After(now.Add(-pendingMaxAge))
	case StatusRunning:
		start := s.CreatedAt.UTC()
		if s.StartedAt != nil {
			start = s.StartedAt.UTC()
		}
		deadline := start.Add(s.ExecutionBudget()).Add(grace)
		return !deadline.After(now)
	default:
		return false
	}
}

func (s *StepRun) ExecutionBudget() time.Duration {
	timeoutMS := s.Timeout
	if timeoutMS <= 0 {
		timeoutMS = DefaultHTTPTimeoutMS
	}

	attempts := 1
	if s.RetryOnFailure {
		attempts = s.RetryCount
		if attempts < 1 {
			attempts = 1
		}
	}

	budget := time.Duration(timeoutMS*attempts) * time.Millisecond
	if s.RetryOnFailure && attempts > 1 && s.RetryDelay > 0 {
		budget += time.Duration(s.RetryDelay*(attempts-1)) * time.Millisecond
	}
	return budget
}
