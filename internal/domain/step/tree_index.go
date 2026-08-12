package step

import (
	"sort"

	"github.com/google/uuid"
)

type GraphEdge struct {
	SourceStepID uuid.UUID
	TargetStepID uuid.UUID
}

func CalculateTreeIndices(
	executionOrderByStepID map[uuid.UUID]int,
	edges []GraphEdge,
) map[uuid.UUID]int {
	if len(executionOrderByStepID) == 0 {
		return map[uuid.UUID]int{}
	}

	parent := make(map[uuid.UUID]uuid.UUID, len(executionOrderByStepID))
	for stepID := range executionOrderByStepID {
		parent[stepID] = stepID
	}

	var find func(uuid.UUID) uuid.UUID
	find = func(id uuid.UUID) uuid.UUID {
		if parent[id] != id {
			parent[id] = find(parent[id])
		}
		return parent[id]
	}

	union := func(a, b uuid.UUID) {
		ra, rb := find(a), find(b)
		if ra != rb {
			parent[ra] = rb
		}
	}

	for _, edge := range edges {
		if _, ok := parent[edge.SourceStepID]; !ok {
			continue
		}
		if _, ok := parent[edge.TargetStepID]; !ok {
			continue
		}
		union(edge.SourceStepID, edge.TargetStepID)
	}

	type component struct {
		minOrder int
		stepIDs  []uuid.UUID
	}

	components := make(map[uuid.UUID]*component)
	for stepID, order := range executionOrderByStepID {
		root := find(stepID)
		comp, ok := components[root]
		if !ok {
			components[root] = &component{
				minOrder: order,
				stepIDs:  []uuid.UUID{stepID},
			}
			continue
		}

		comp.stepIDs = append(comp.stepIDs, stepID)
		if order < comp.minOrder {
			comp.minOrder = order
		}
	}

	sorted := make([]*component, 0, len(components))
	for _, c := range components {
		sorted = append(sorted, c)
	}

	sort.Slice(sorted, func(i, j int) bool {
		if len(sorted[i].stepIDs) != len(sorted[j].stepIDs) {
			return len(sorted[i].stepIDs) > len(sorted[j].stepIDs)
		}
		return sorted[i].minOrder < sorted[j].minOrder
	})

	treeIndexByStepID := make(map[uuid.UUID]int, len(executionOrderByStepID))
	for treeIndex, comp := range sorted {
		for _, stepID := range comp.stepIDs {
			treeIndexByStepID[stepID] = treeIndex + 1
		}
	}

	return treeIndexByStepID
}
