package step

import (
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type PositionedStep struct {
	ID        uuid.UUID
	Position  Position
	CreatedAt time.Time
}

type StepOrdering struct {
	Index          string
	ExecutionOrder int
}

func CalculateOrderingByPosition(steps []PositionedStep, edges []GraphEdge) map[uuid.UUID]StepOrdering {
	if len(steps) == 0 {
		return map[uuid.UUID]StepOrdering{}
	}

	byID := make(map[uuid.UUID]PositionedStep, len(steps))
	inDegree := make(map[uuid.UUID]int, len(steps))
	children := make(map[uuid.UUID][]uuid.UUID, len(steps))
	for _, step := range steps {
		byID[step.ID] = step
		inDegree[step.ID] = 0
		children[step.ID] = []uuid.UUID{}
	}

	// At a split, prefer leftmost then topmost branch.
	compare := func(a, b uuid.UUID) bool {
		sa := byID[a]
		sb := byID[b]
		if sa.Position.X != sb.Position.X {
			return sa.Position.X < sb.Position.X
		}
		if sa.Position.Y != sb.Position.Y {
			return sa.Position.Y < sb.Position.Y
		}
		if !sa.CreatedAt.Equal(sb.CreatedAt) {
			return sa.CreatedAt.Before(sb.CreatedAt)
		}
		return a.String() < b.String()
	}

	for _, edge := range edges {
		if _, ok := byID[edge.SourceStepID]; !ok {
			continue
		}
		if _, ok := byID[edge.TargetStepID]; !ok {
			continue
		}
		children[edge.SourceStepID] = append(children[edge.SourceStepID], edge.TargetStepID)
		inDegree[edge.TargetStepID]++
	}

	for stepID := range children {
		targets := children[stepID]
		sort.Slice(targets, func(i, j int) bool {
			return compare(targets[i], targets[j])
		})
		children[stepID] = targets
	}

	all := make([]uuid.UUID, 0, len(steps))
	for stepID := range byID {
		all = append(all, stepID)
	}
	sort.Slice(all, func(i, j int) bool {
		return compare(all[i], all[j])
	})

	roots := make([]uuid.UUID, 0, len(all))
	for _, stepID := range all {
		if inDegree[stepID] == 0 {
			roots = append(roots, stepID)
		}
	}

	ordered := make([]uuid.UUID, 0, len(all))
	visited := make(map[uuid.UUID]bool, len(all))
	var dfs func(id uuid.UUID)
	dfs = func(id uuid.UUID) {
		if visited[id] {
			return
		}
		visited[id] = true
		ordered = append(ordered, id)
		for _, childID := range children[id] {
			dfs(childID)
		}
	}

	for _, rootID := range roots {
		dfs(rootID)
	}
	for _, stepID := range all {
		dfs(stepID)
	}

	ordering := make(map[uuid.UUID]StepOrdering, len(ordered))
	for i, stepID := range ordered {
		index := strconv.Itoa(i + 1)
		ordering[stepID] = StepOrdering{
			Index:          index,
			ExecutionOrder: CalculateExecutionOrder(index),
		}
	}
	return ordering
}
