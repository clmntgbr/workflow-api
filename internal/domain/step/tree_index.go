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

	adj := make(map[uuid.UUID][]uuid.UUID, len(executionOrderByStepID))
	inDegree := make(map[uuid.UUID]int, len(executionOrderByStepID))
	for stepID := range executionOrderByStepID {
		adj[stepID] = []uuid.UUID{}
		inDegree[stepID] = 0
	}
	for _, edge := range edges {
		if _, ok := executionOrderByStepID[edge.SourceStepID]; !ok {
			continue
		}
		if _, ok := executionOrderByStepID[edge.TargetStepID]; !ok {
			continue
		}
		adj[edge.SourceStepID] = append(adj[edge.SourceStepID], edge.TargetStepID)
		inDegree[edge.TargetStepID]++
	}

	compareStepIDs := func(a, b uuid.UUID) bool {
		oa := executionOrderByStepID[a]
		ob := executionOrderByStepID[b]
		if oa != ob {
			return oa < ob
		}
		return a.String() < b.String()
	}

	for sourceID := range adj {
		targets := adj[sourceID]
		sort.Slice(targets, func(i, j int) bool {
			return compareStepIDs(targets[i], targets[j])
		})
		adj[sourceID] = targets
	}

	allSteps := make([]uuid.UUID, 0, len(executionOrderByStepID))
	for stepID := range executionOrderByStepID {
		allSteps = append(allSteps, stepID)
	}
	sort.Slice(allSteps, func(i, j int) bool {
		return compareStepIDs(allSteps[i], allSteps[j])
	})

	roots := make([]uuid.UUID, 0, len(allSteps))
	for _, stepID := range allSteps {
		if inDegree[stepID] == 0 {
			roots = append(roots, stepID)
		}
	}

	treeIndexByStepID := make(map[uuid.UUID]int, len(executionOrderByStepID))
	nextBranchID := 1

	var assign func(stepID uuid.UUID, branchID int)
	assign = func(stepID uuid.UUID, branchID int) {
		if _, exists := treeIndexByStepID[stepID]; exists {
			return
		}

		treeIndexByStepID[stepID] = branchID

		targets := adj[stepID]
		for i, targetID := range targets {
			childBranchID := branchID
			if i > 0 {
				nextBranchID++
				childBranchID = nextBranchID
			}
			assign(targetID, childBranchID)
		}
	}

	for _, rootID := range roots {
		if _, exists := treeIndexByStepID[rootID]; exists {
			continue
		}
		assign(rootID, nextBranchID)
		nextBranchID++
	}

	for _, stepID := range allSteps {
		if _, exists := treeIndexByStepID[stepID]; exists {
			continue
		}
		assign(stepID, nextBranchID)
		nextBranchID++
	}

	return treeIndexByStepID
}
