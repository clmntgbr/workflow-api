package workflowrun

import (
	"context"
	"errors"
	"fmt"

	domaincondition "go-api/internal/domain/condition"
	domainconnection "go-api/internal/domain/connection"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainvariable "go-api/internal/domain/variable"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

func (h *Orchestrator) activateStepRun(
	ctx context.Context,
	run *domainworkflowrun.WorkflowRun,
	step domainstep.StepView,
	stepsByID map[uuid.UUID]domainstep.StepView,
	connections []domainconnection.ConnectionView,
	runsByStep map[uuid.UUID]*domainsteprun.StepRun,
) ([]*domainsteprun.StepRun, error) {
	if _, exists := runsByStep[step.ID]; exists {
		return nil, nil
	}
	if isOrphanDelay(step, connections) || isOrphanCondition(step, connections) {
		return nil, nil
	}
	if !canEnqueue(step.ID, connections, runsByStep) {
		return nil, nil
	}

	stepRun, err := h.buildStepRun(ctx, run, step)
	if err != nil {
		return nil, err
	}
	if err := h.stepRunRepo.Save(ctx, stepRun); err != nil {
		return nil, err
	}
	runsByStep[step.ID] = stepRun

	changed := []*domainsteprun.StepRun{stepRun}
	if step.Type == domainstep.TypeCondition {
		more, err := h.executeConditionStepRun(ctx, run, step, stepRun, stepsByID, connections, runsByStep)
		if err != nil {
			return nil, err
		}
		changed = append(changed, more...)
	}
	return changed, nil
}

func (h *Orchestrator) executeConditionStepRun(
	ctx context.Context,
	run *domainworkflowrun.WorkflowRun,
	step domainstep.StepView,
	stepRun *domainsteprun.StepRun,
	stepsByID map[uuid.UUID]domainstep.StepView,
	connections []domainconnection.ConnectionView,
	runsByStep map[uuid.UUID]*domainsteprun.StepRun,
) ([]*domainsteprun.StepRun, error) {
	changed := make([]*domainsteprun.StepRun, 0)

	if err := stepRun.MarkStarted(); err != nil {
		return nil, err
	}
	if err := h.stepRunRepo.Update(ctx, stepRun); err != nil {
		return nil, err
	}

	outgoing := outgoingConnections(step.ID, connections)
	if len(outgoing) != 2 {
		if err := stepRun.MarkFailed("condition step must have exactly two outgoing branches", nil, nil); err != nil {
			return nil, err
		}
		if err := h.stepRunRepo.Update(ctx, stepRun); err != nil {
			return nil, err
		}
		skipped, err := h.skipConditionBranchTargets(ctx, run, outgoing, stepsByID, connections, runsByStep)
		if err != nil {
			return nil, err
		}
		return append(changed, skipped...), nil
	}

	expression := ""
	if step.Expression != nil {
		expression = *step.Expression
	}

	matched, evalErr := domaincondition.EvaluateBoolean(expression, run.Context)
	if evalErr != nil {
		message := evalErr.Error()
		var missing *domainvariable.MissingVariableError
		if errors.As(evalErr, &missing) {
			message = fmt.Sprintf("variable %s not found", missing.Key)
		}
		if err := stepRun.MarkFailed(message, nil, nil); err != nil {
			return nil, err
		}
		if err := h.stepRunRepo.Update(ctx, stepRun); err != nil {
			return nil, err
		}
		skipped, err := h.skipConditionBranchTargets(ctx, run, outgoing, stepsByID, connections, runsByStep)
		if err != nil {
			return nil, err
		}
		return append(changed, skipped...), nil
	}

	if err := stepRun.MarkConditionSucceeded(matched); err != nil {
		return nil, err
	}
	if err := h.stepRunRepo.Update(ctx, stepRun); err != nil {
		return nil, err
	}

	for _, conn := range outgoing {
		if conn.Branch == nil {
			continue
		}
		target, ok := stepsByID[conn.TargetStepID]
		if !ok {
			continue
		}
		branchIsTrue := *conn.Branch == domainconnection.ConditionBranchTrue
		if branchIsTrue == matched {
			more, err := h.activateStepRun(ctx, run, target, stepsByID, connections, runsByStep)
			if err != nil {
				return nil, err
			}
			changed = append(changed, more...)
			continue
		}
		skipped, err := h.skipBranch(ctx, run, target, stepsByID, connections, runsByStep)
		if err != nil {
			return nil, err
		}
		changed = append(changed, skipped...)
	}

	return changed, nil
}

func (h *Orchestrator) skipConditionBranchTargets(
	ctx context.Context,
	run *domainworkflowrun.WorkflowRun,
	outgoing []domainconnection.ConnectionView,
	stepsByID map[uuid.UUID]domainstep.StepView,
	connections []domainconnection.ConnectionView,
	runsByStep map[uuid.UUID]*domainsteprun.StepRun,
) ([]*domainsteprun.StepRun, error) {
	changed := make([]*domainsteprun.StepRun, 0)
	for _, conn := range outgoing {
		target, ok := stepsByID[conn.TargetStepID]
		if !ok {
			continue
		}
		skipped, err := h.skipBranch(ctx, run, target, stepsByID, connections, runsByStep)
		if err != nil {
			return nil, err
		}
		changed = append(changed, skipped...)
	}
	return changed, nil
}

func outgoingConnections(
	stepID uuid.UUID,
	connections []domainconnection.ConnectionView,
) []domainconnection.ConnectionView {
	out := make([]domainconnection.ConnectionView, 0)
	for _, conn := range connections {
		if conn.SourceStepID == stepID {
			out = append(out, conn)
		}
	}
	return out
}
