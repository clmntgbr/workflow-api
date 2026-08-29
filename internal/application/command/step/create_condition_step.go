package step

import (
	"context"
	"errors"
	"strings"

	cmdquota "go-api/internal/application/command/quota"
	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/application/messaging"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainworkflow "go-api/internal/domain/workflow"
	"time"

	"github.com/google/uuid"
)

type CreateConditionStepCommand struct {
	UserID     uuid.UUID
	WorkflowID uuid.UUID
	ProjectID  uuid.UUID
	Name       string
	Expression string
	Position   domainstep.Position
}

type CreateConditionStepHandler struct {
	stepRepo     domainstep.StepWriteRepository
	stepReadRepo domainstep.StepReadRepository
	connReadRepo domainconnection.ConnectionReadRepository
	workflowRepo domainworkflow.WorkflowReadRepository
	outbox       port.OutboxRepository
	assert       *cmdquota.AssertCreateAllowedHandler
}

func NewCreateConditionStepHandler(
	stepRepo domainstep.StepWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	workflowRepo domainworkflow.WorkflowReadRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *CreateConditionStepHandler {
	return &CreateConditionStepHandler{
		stepRepo:     stepRepo,
		stepReadRepo: stepReadRepo,
		connReadRepo: connReadRepo,
		workflowRepo: workflowRepo,
		outbox:       outbox,
		assert:       assert,
	}
}

func (h *CreateConditionStepHandler) Handle(
	ctx context.Context,
	cmd CreateConditionStepCommand,
) (*domainstep.Step, error) {
	if cmd.WorkflowID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}
	if cmd.ProjectID == uuid.Nil {
		return nil, errors.New("projectId is required")
	}
	if cmd.UserID == uuid.Nil {
		return nil, errors.New("userId is required")
	}
	if strings.TrimSpace(cmd.Expression) == "" {
		return nil, domainstep.ErrInvalidStepTypeConfig
	}

	workflow, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to get workflow")
	}
	if workflow == nil || workflow.Status == domainworkflow.StatusDeleted {
		return nil, errors.New("workflow not found")
	}
	if workflow.ProjectID != cmd.ProjectID {
		return nil, errors.New("workflow not found")
	}

	if err := h.assert.AssertStepCreate(ctx, cmd.UserID, cmd.ProjectID, cmd.WorkflowID); err != nil {
		return nil, err
	}

	now := domainstep.PositionedStep{
		ID:        uuid.New(),
		Position:  cmd.Position,
		CreatedAt: time.Now().UTC(),
	}
	existingSteps, err := h.stepReadRepo.FindByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list steps")
	}
	positioned := make([]domainstep.PositionedStep, 0, len(existingSteps)+1)
	for _, stepView := range existingSteps {
		positioned = append(positioned, domainstep.PositionedStep{
			ID:        stepView.ID,
			Position:  stepView.Position,
			CreatedAt: stepView.CreatedAt,
		})
	}
	positioned = append(positioned, now)

	connections, err := h.connReadRepo.FindByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to list connections")
	}
	edges := make([]domainstep.GraphEdge, 0, len(connections))
	for _, connection := range connections {
		edges = append(edges, domainstep.GraphEdge{
			SourceStepID: connection.SourceStepID,
			TargetStepID: connection.TargetStepID,
		})
	}
	ordering := domainstep.CalculateOrderingByPosition(positioned, edges)

	executionOrderByStepID := make(map[uuid.UUID]int, len(ordering))
	for stepID, values := range ordering {
		executionOrderByStepID[stepID] = values.ExecutionOrder
	}
	treeIndices := domainstep.CalculateTreeIndices(executionOrderByStepID, edges)

	s, err := domainstep.NewConditionStep(domainstep.NewConditionStepParams{
		ID:             now.ID,
		WorkflowID:     cmd.WorkflowID,
		ProjectID:      cmd.ProjectID,
		Name:           strings.TrimSpace(cmd.Name),
		Expression:     cmd.Expression,
		Index:          ordering[now.ID].Index,
		ExecutionOrder: ordering[now.ID].ExecutionOrder,
		TreeIndex:      treeIndices[now.ID],
		Position:       cmd.Position,
	})
	if err != nil {
		return nil, err
	}

	err = h.stepRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.stepRepo.Save(txCtx, s); err != nil {
			return err
		}
		if err := h.stepRepo.UpdateOrdering(txCtx, ordering); err != nil {
			return err
		}
		if err := h.stepRepo.UpdateTreeIndices(txCtx, treeIndices); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(s.PullEvents(), cmd.UserID))
	})
	if err != nil {
		return nil, errors.New("failed to create step")
	}

	return s, nil
}
