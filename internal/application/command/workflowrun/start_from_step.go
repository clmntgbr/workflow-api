package workflowrun

import (
	"context"
	"sort"

	domaininsight "go-api/internal/domain/insight"
	domainstep "go-api/internal/domain/step"
	domainsteprun "go-api/internal/domain/steprun"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

type startFromReplay struct {
	copies   []*domainsteprun.StepRun
	skips    []*domainsteprun.StepRun
	insights []*domaininsight.Insight
}

func (h *StartWorkflowRunHandler) prepareStartFrom(
	ctx context.Context,
	fromStepID uuid.UUID,
	run *domainworkflowrun.WorkflowRun,
) (*startFromReplay, error) {
	steps, err := h.stepRead.FindByWorkflowID(ctx, run.WorkflowID)
	if err != nil {
		return nil, err
	}
	connections, err := h.connRead.FindByWorkflowID(ctx, run.WorkflowID)
	if err != nil {
		return nil, err
	}

	var fromStep *domainstep.StepView
	stepsByID := make(map[uuid.UUID]domainstep.StepView, len(steps))
	for i := range steps {
		stepsByID[steps[i].ID] = steps[i]
		if steps[i].ID == fromStepID {
			fromStep = &steps[i]
		}
	}
	if fromStep == nil ||
		isOrphanDelay(*fromStep, connections) ||
		isOrphanCondition(*fromStep, connections) {
		return nil, domainworkflowrun.ErrFromStepNotFound
	}

	ancestors := ancestorIDs(fromStepID, connections)
	descendants := descendantIDs(fromStepID, connections)
	keep := make(map[uuid.UUID]struct{}, len(ancestors)+len(descendants)+1)
	keep[fromStepID] = struct{}{}
	for _, id := range ancestors {
		keep[id] = struct{}{}
	}
	for id := range descendants {
		keep[id] = struct{}{}
	}

	replay := &startFromReplay{}
	if len(ancestors) > 0 {
		sort.Slice(ancestors, func(i, j int) bool {
			return stepsByID[ancestors[i]].ExecutionOrder < stepsByID[ancestors[j]].ExecutionOrder
		})
		for _, stepID := range ancestors {
			latest, err := h.stepRunRead.FindLatestCompletedByStepID(ctx, stepID)
			if err != nil {
				return nil, err
			}
			if latest == nil {
				return nil, domainworkflowrun.ErrMissingPreviousStepRun
			}
			cloned := latest.ReplayOnto(run.ID)
			replay.copies = append(replay.copies, cloned)
			run.MergeContext(cloned.ExtractedVariables)

			sourceInsights, err := h.insightRead.FindByStepRunID(ctx, latest.ID)
			if err != nil {
				return nil, err
			}
			for _, insight := range sourceInsights {
				replay.insights = append(replay.insights, insight.ReplayOnto(cloned.ID))
			}
		}
	}

	for stepID := range reachableStepIDs(steps, connections) {
		if _, ok := keep[stepID]; ok {
			continue
		}
		step, ok := stepsByID[stepID]
		if !ok {
			continue
		}
		skipped := domainsteprun.NewStepRun(domainsteprun.NewStepRunParams{
			WorkflowRunID:        run.ID,
			StepID:               step.ID,
			WorkflowID:           step.WorkflowID,
			ProjectID:            step.ProjectID,
			StepType:             step.Type,
			DelayDurationSeconds: step.DelayDurationSeconds,
			Name:                 step.Name,
			Description:          step.Description,
			Index:                step.Index,
			ExecutionOrder:       step.ExecutionOrder,
			TreeIndex:            step.TreeIndex,
			Position:             step.Position,
		})
		if err := skipped.MarkSkipped(); err != nil {
			return nil, err
		}
		replay.skips = append(replay.skips, skipped)
	}

	return replay, nil
}
