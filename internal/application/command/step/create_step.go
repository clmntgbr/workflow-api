package step

import (
	"context"
	"errors"

	domainendpoint "go-api/internal/domain/endpoint"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type CreateStepCommand struct {
	WorkflowID     uuid.UUID
	EndpointID     uuid.UUID
	OrganizationID uuid.UUID
	Index          string
	TreeIndex      int
	Position       domainstep.Position
}

type CreateStepHandler struct {
	stepRepo     domainstep.StepWriteRepository
	endpointRepo domainendpoint.EndpointReadRepository
	workflowRepo domainworkflow.WorkflowReadRepository
	outbox       port.OutboxRepository
}

func NewCreateStepHandler(
	stepRepo domainstep.StepWriteRepository,
	endpointRepo domainendpoint.EndpointReadRepository,
	workflowRepo domainworkflow.WorkflowReadRepository,
	outbox port.OutboxRepository,
) *CreateStepHandler {
	return &CreateStepHandler{
		stepRepo:     stepRepo,
		endpointRepo: endpointRepo,
		workflowRepo: workflowRepo,
		outbox:       outbox,
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
	if cmd.OrganizationID == uuid.Nil {
		return nil, errors.New("organizationId is required")
	}
	if cmd.Index == "" {
		return nil, errors.New("index is required")
	}

	workflow, err := h.workflowRepo.FindByID(ctx, cmd.WorkflowID)
	if err != nil {
		return nil, errors.New("failed to get workflow")
	}
	if workflow == nil || workflow.Status == domainworkflow.StatusDeleted {
		return nil, errors.New("workflow not found")
	}
	if workflow.OrganizationID != cmd.OrganizationID {
		return nil, errors.New("workflow not found")
	}

	endpoint, err := h.endpointRepo.FindByID(ctx, cmd.EndpointID)
	if err != nil {
		return nil, errors.New("failed to get endpoint")
	}
	if endpoint == nil || endpoint.Status == domainendpoint.StatusDeleted {
		return nil, errors.New("endpoint not found")
	}
	if endpoint.OrganizationID != cmd.OrganizationID {
		return nil, errors.New("endpoint not found")
	}

	var created *domainstep.Step
	err = h.stepRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		executionOrder, err := h.stepRepo.NextExecutionOrder(txCtx, cmd.WorkflowID)
		if err != nil {
			return errors.New("failed to compute execution order")
		}

		s := domainstep.NewStep(domainstep.NewStepParams{
			WorkflowID:     cmd.WorkflowID,
			EndpointID:     cmd.EndpointID,
			OrganizationID: cmd.OrganizationID,
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
			Index:          cmd.Index,
			ExecutionOrder: executionOrder,
			TreeIndex:      cmd.TreeIndex,
			Position:       cmd.Position,
		})
		created = s

		if err := h.stepRepo.Save(txCtx, s); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, s.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to create step")
	}

	return created, nil
}
