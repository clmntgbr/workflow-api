package connection

import (
	"errors"

	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

var (
	ErrInvalidBranch                    = errors.New("invalid connection branch")
	ErrConditionRequiresBranch          = errors.New("condition step connections require a branch")
	ErrNonConditionBranchForbidden      = errors.New("branch is only allowed on condition step connections")
	ErrConditionOutgoingCount           = errors.New("condition steps must have exactly two outgoing connections with true and false branches")
	ErrConditionalTargetMultipleParents = errors.New("steps reached via a conditional branch cannot have more than one incoming connection")
)

func ValidateWorkflowConnections(
	steps []domainstep.StepView,
	connections []ConnectionView,
) error {
	stepsByID := make(map[uuid.UUID]domainstep.StepView, len(steps))
	for _, step := range steps {
		stepsByID[step.ID] = step
	}

	incomingCount := make(map[uuid.UUID]int, len(steps))
	for _, conn := range connections {
		incomingCount[conn.TargetStepID]++
	}

	for _, conn := range connections {
		source, ok := stepsByID[conn.SourceStepID]
		if !ok {
			continue
		}

		if source.Type != domainstep.TypeCondition {
			if conn.Branch != nil {
				return ErrNonConditionBranchForbidden
			}
			continue
		}

		if conn.Branch == nil {
			return ErrConditionRequiresBranch
		}
		if !conn.Branch.Valid() {
			return ErrInvalidBranch
		}

		if incomingCount[conn.TargetStepID] != 1 {
			return ErrConditionalTargetMultipleParents
		}
	}

	outgoingBySource := make(map[uuid.UUID][]ConnectionView, len(steps))
	for _, conn := range connections {
		outgoingBySource[conn.SourceStepID] = append(outgoingBySource[conn.SourceStepID], conn)
	}

	for _, step := range steps {
		if step.Type != domainstep.TypeCondition {
			continue
		}

		outgoing := outgoingBySource[step.ID]
		switch len(outgoing) {
		case 0, 1:
			continue
		case 2:
			var hasTrue, hasFalse bool
			for _, conn := range outgoing {
				if conn.Branch == nil {
					return ErrConditionRequiresBranch
				}
				switch *conn.Branch {
				case ConditionBranchTrue:
					hasTrue = true
				case ConditionBranchFalse:
					hasFalse = true
				default:
					return ErrInvalidBranch
				}
			}
			if !hasTrue || !hasFalse {
				return ErrConditionOutgoingCount
			}
		default:
			return ErrConditionOutgoingCount
		}
	}

	return nil
}
