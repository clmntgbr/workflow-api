package workflowrun

import (
	domainconnection "go-api/internal/domain/connection"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

func incomingIDs(stepID uuid.UUID, connections []domainconnection.ConnectionView) []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	for _, conn := range connections {
		if conn.TargetStepID == stepID {
			ids = append(ids, conn.SourceStepID)
		}
	}
	return ids
}

func outgoingIDs(stepID uuid.UUID, connections []domainconnection.ConnectionView) []uuid.UUID {
	ids := make([]uuid.UUID, 0)
	for _, conn := range connections {
		if conn.SourceStepID == stepID {
			ids = append(ids, conn.TargetStepID)
		}
	}
	return ids
}

func isOrphanDelay(step domainstep.StepView, connections []domainconnection.ConnectionView) bool {
	if step.Type != domainstep.TypeDelay {
		return false
	}
	return len(incomingIDs(step.ID, connections)) == 0 &&
		len(outgoingIDs(step.ID, connections)) == 0
}

func isOrphanCondition(step domainstep.StepView, connections []domainconnection.ConnectionView) bool {
	if step.Type != domainstep.TypeCondition {
		return false
	}
	return len(incomingIDs(step.ID, connections)) == 0 &&
		len(outgoingIDs(step.ID, connections)) == 0
}

func ancestorIDs(from uuid.UUID, connections []domainconnection.ConnectionView) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{})
	queue := incomingIDs(from, connections)
	for _, id := range queue {
		seen[id] = struct{}{}
	}
	for i := 0; i < len(queue); i++ {
		for _, parent := range incomingIDs(queue[i], connections) {
			if _, ok := seen[parent]; ok {
				continue
			}
			seen[parent] = struct{}{}
			queue = append(queue, parent)
		}
	}
	return queue
}

func descendantIDs(from uuid.UUID, connections []domainconnection.ConnectionView) map[uuid.UUID]struct{} {
	seen := make(map[uuid.UUID]struct{})
	queue := outgoingIDs(from, connections)
	for _, id := range queue {
		seen[id] = struct{}{}
	}
	for i := 0; i < len(queue); i++ {
		for _, next := range outgoingIDs(queue[i], connections) {
			if _, ok := seen[next]; ok {
				continue
			}
			seen[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	return seen
}

func reachableStepIDs(
	steps []domainstep.StepView,
	connections []domainconnection.ConnectionView,
) map[uuid.UUID]struct{} {
	hasIncoming := make(map[uuid.UUID]bool, len(steps))
	for _, conn := range connections {
		hasIncoming[conn.TargetStepID] = true
	}

	seen := make(map[uuid.UUID]struct{}, len(steps))
	queue := make([]uuid.UUID, 0)
	for _, step := range steps {
		if hasIncoming[step.ID] {
			continue
		}
		if isOrphanDelay(step, connections) || isOrphanCondition(step, connections) {
			continue
		}
		queue = append(queue, step.ID)
		seen[step.ID] = struct{}{}
	}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range outgoingIDs(current, connections) {
			if _, ok := seen[next]; ok {
				continue
			}
			seen[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	return seen
}
