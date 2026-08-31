package rabbitmq_test

import (
	"testing"

	"go-api/internal/infrastructure/messaging/rabbitmq"
)

func TestExecutorTopology_EnablesQueuePriority(t *testing.T) {
	topology := rabbitmq.ExecutorTopology("step_run.execute", "step_run.execute", "step_run.execute", 30000)
	if topology.MaxPriority != rabbitmq.MaxExecutorQueuePriority {
		t.Fatalf("max priority: got %d want %d", topology.MaxPriority, rabbitmq.MaxExecutorQueuePriority)
	}
}

func TestDefaultTopology_DoesNotEnableQueuePriority(t *testing.T) {
	topology := rabbitmq.DefaultTopology("events", "events", "events", 30000)
	if topology.MaxPriority != 0 {
		t.Fatalf("max priority: got %d want 0", topology.MaxPriority)
	}
}
