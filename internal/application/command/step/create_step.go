package step

import (
	"context"
	"errors"
	cmdquota "go-api/internal/application/command/quota"
	domainconnection "go-api/internal/domain/connection"
	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainworkflow "go-api/internal/domain/workflow"
	"time"

	"github.com/google/uuid"
)

type CreateStepCommand struct {
	UserID     uuid.UUID
	WorkflowID uuid.UUID
	EndpointID uuid.UUID
	ProjectID  uuid.UUID
	Position   domainstep.Position
}

type CreateStepHandler struct {
	stepRepo     domainstep.StepWriteRepository
	stepReadRepo domainstep.StepReadRepository
	connReadRepo domainconnection.ConnectionReadRepository
	endpointRepo domainendpoint.EndpointReadRepository
	workflowRepo domainworkflow.WorkflowReadRepository
	outbox       port.OutboxRepository
	assert       *cmdquota.AssertCreateAllowedHandler
}

func NewCreateStepHandler(
	stepRepo domainstep.StepWriteRepository,
	stepReadRepo domainstep.StepReadRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	endpointRepo domainendpoint.EndpointReadRepository,
	workflowRepo domainworkflow.WorkflowReadRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *CreateStepHandler {
	return &CreateStepHandler{
		stepRepo:     stepRepo,
		stepReadRepo: stepReadRepo,
		connReadRepo: connReadRepo,
		endpointRepo: endpointRepo,
		workflowRepo: workflowRepo,
		outbox:       outbox,
		assert:       assert,
	}
}

func (h *CreateStepHandler) Handle(
	ctx context.Context,
	cmd CreateStepCommand,
) (*domainstep.Step, error) {
	if cmd.WorkflowID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}
	if cmd.EndpointID == uuid.Nil {
		return nil, errors.New("endpointId is required")
	}
	if cmd.ProjectID == uuid.Nil {
		return nil, errors.New("projectId is required")
	}
	if cmd.UserID == uuid.Nil {
		return nil, errors.New("userId is required")
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

	endpoint, err := h.endpointRepo.FindByID(ctx, cmd.EndpointID)
	if err != nil {
		return nil, errors.New("failed to get endpoint")
	}
	if endpoint == nil || endpoint.Status == domainendpoint.StatusDeleted {
		return nil, errors.New("endpoint not found")
	}
	if endpoint.ProjectID != cmd.ProjectID {
		return nil, errors.New("endpoint not found")
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

	s := domainstep.NewStep(domainstep.NewStepParams{
		ID:         now.ID,
		WorkflowID: cmd.WorkflowID,
		EndpointID: cmd.EndpointID,
		ProjectID:  cmd.ProjectID,
		Endpoint: domainstep.EndpointSnapshot{
			ID:             endpoint.ID,
			Name:           endpoint.Name,
			Description:    endpoint.Description,
			URL:            endpoint.URL,
			Method:         string(endpoint.Method),
			Headers:        endpoint.Headers,
			Query:          endpoint.Query,
			Body:           endpoint.Body,
			Timeout:        endpoint.Timeout,
			RetryOnFailure: endpoint.RetryOnFailure,
			RetryCount:     endpoint.RetryCount,
			RetryDelay:     endpoint.RetryDelay,
		},
		Index:          ordering[now.ID].Index,
		ExecutionOrder: ordering[now.ID].ExecutionOrder,
		TreeIndex:      treeIndices[now.ID],
		Position:       cmd.Position,
	})

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
		return h.outbox.StoreEvents(txCtx, s.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create step")
	}

	return s, nil
}
