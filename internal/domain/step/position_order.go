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

func CalculateOrderingByPosition(steps []PositionedStep) map[uuid.UUID]StepOrdering {
	sorted := make([]PositionedStep, len(steps))
	copy(sorted, steps)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Position.Y != sorted[j].Position.Y {
			return sorted[i].Position.Y < sorted[j].Position.Y
		}
		if sorted[i].Position.X != sorted[j].Position.X {
			return sorted[i].Position.X < sorted[j].Position.X
		}
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	ordering := make(map[uuid.UUID]StepOrdering, len(sorted))
	for i, step := range sorted {
		index := strconv.Itoa(i + 1)
		ordering[step.ID] = StepOrdering{
			Index:          index,
			ExecutionOrder: CalculateExecutionOrder(index),
		}
	}
	return ordering
}
