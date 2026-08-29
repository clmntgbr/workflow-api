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

func isOrphanDelay(
	step domainstep.StepView,
	connections []domainconnection.ConnectionView,
) bool {
	if step.Type != domainstep.TypeDelay {
		return false
	}
	return len(incomingIDs(step.ID, connections)) == 0 &&
		len(outgoingIDs(step.ID, connections)) == 0
}

func isOrphanCondition(
	step domainstep.StepView,
	connections []domainconnection.ConnectionView,
) bool {
	if step.Type != domainstep.TypeCondition {
		return false
	}
	return len(incomingIDs(step.ID, connections)) == 0 &&
		len(outgoingIDs(step.ID, connections)) == 0
}

func rootSteps(
	steps []domainstep.StepView,
	connections []domainconnection.ConnectionView,
) []domainstep.StepView {
	hasIncoming := make(map[uuid.UUID]bool, len(steps))
	for _, conn := range connections {
		hasIncoming[conn.TargetStepID] = true
	}

	roots := make([]domainstep.StepView, 0)
	for _, step := range steps {
		if !hasIncoming[step.ID] {
			roots = append(roots, step)
		}
	}
	return roots
}

func stepByID(steps []domainstep.StepView) map[uuid.UUID]domainstep.StepView {
	out := make(map[uuid.UUID]domainstep.StepView, len(steps))
	for _, step := range steps {
		out[step.ID] = step
	}
	return out
}

func reachableStepIDs(
	steps []domainstep.StepView,
	connections []domainconnection.ConnectionView,
) map[uuid.UUID]struct{} {
	roots := rootSteps(steps, connections)
	seen := make(map[uuid.UUID]struct{}, len(steps))
	queue := make([]uuid.UUID, 0, len(roots))
	for _, root := range roots {
		if isOrphanDelay(root, connections) || isOrphanCondition(root, connections) {
			continue
		}
		queue = append(queue, root.ID)
		seen[root.ID] = struct{}{}
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
