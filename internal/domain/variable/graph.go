package variable

import (
	"github.com/google/uuid"
)

type GraphEdge struct {
	SourceStepID uuid.UUID
	TargetStepID uuid.UUID
}

func AncestorStepIDs(stepID uuid.UUID, edges []GraphEdge) map[uuid.UUID]struct{} {
	incoming := map[uuid.UUID][]uuid.UUID{}
	for _, edge := range edges {
		incoming[edge.TargetStepID] = append(incoming[edge.TargetStepID], edge.SourceStepID)
	}

	ancestors := map[uuid.UUID]struct{}{}
	queue := append([]uuid.UUID{}, incoming[stepID]...)
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if _, ok := ancestors[current]; ok {
			continue
		}
		ancestors[current] = struct{}{}
		queue = append(queue, incoming[current]...)
	}
	return ancestors
}

func ValidateReferences(
	stepID uuid.UUID,
	referencedIDs []uuid.UUID,
	variablesByID map[uuid.UUID]VariableView,
	edges []GraphEdge,
) error {
	ancestors := AncestorStepIDs(stepID, edges)
	for _, id := range referencedIDs {
		variable, ok := variablesByID[id]
		if !ok {
			return ErrInvalidRef
		}
		if _, ok := ancestors[variable.StepID]; !ok {
			return ErrNotAncestor
		}
	}
	return nil
}
