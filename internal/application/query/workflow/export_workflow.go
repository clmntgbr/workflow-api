package workflow

import (
	"context"
	"errors"
	"time"

	"go-api/internal/application/workflowio"
	domainassertion "go-api/internal/domain/assertion"
	domainconnection "go-api/internal/domain/connection"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"
	domainworkflow "go-api/internal/domain/workflow"

	"github.com/google/uuid"
)

type ExportWorkflowQuery struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
}

type ExportWorkflowHandler struct {
	workflows   domainworkflow.WorkflowReadRepository
	steps       domainstep.StepReadRepository
	connections domainconnection.ConnectionReadRepository
	variables   domainvariable.VariableReadRepository
	assertions  domainassertion.AssertionReadRepository
}

func NewExportWorkflowHandler(
	workflows domainworkflow.WorkflowReadRepository,
	steps domainstep.StepReadRepository,
	connections domainconnection.ConnectionReadRepository,
	variables domainvariable.VariableReadRepository,
	assertions domainassertion.AssertionReadRepository,
) *ExportWorkflowHandler {
	return &ExportWorkflowHandler{
		workflows:   workflows,
		steps:       steps,
		connections: connections,
		variables:   variables,
		assertions:  assertions,
	}
}

func (h *ExportWorkflowHandler) Handle(ctx context.Context, q ExportWorkflowQuery) (*workflowio.Document, error) {
	if q.ID == uuid.Nil {
		return nil, errors.New("workflowId is required")
	}
	if q.ProjectID == uuid.Nil {
		return nil, errors.New("projectId is required")
	}

	workflow, err := h.workflows.FindByID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to get workflow")
	}
	if workflow == nil || workflow.Status == domainworkflow.StatusDeleted {
		return nil, errors.New("workflow not found")
	}
	if workflow.ProjectID != q.ProjectID {
		return nil, errors.New("workflow not found")
	}

	steps, err := h.steps.FindByWorkflowID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to list steps")
	}
	connections, err := h.connections.FindByWorkflowID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to list connections")
	}
	variables, err := h.variables.FindByWorkflowID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to list variables")
	}
	assertions, err := h.assertions.FindByWorkflowID(ctx, q.ID)
	if err != nil {
		return nil, errors.New("failed to list assertions")
	}

	doc, err := workflowio.BuildDocument(*workflow, steps, connections, variables, assertions, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return &doc, nil
}
