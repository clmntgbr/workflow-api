package step

import (
	"context"
	"errors"

	domainconnection "go-api/internal/domain/connection"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

type UpdateStepCommand struct {
	ID             uuid.UUID
	WorkflowID     uuid.UUID
	OrganizationID uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         string
	Headers        map[string]string
	Query          map[string]string
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
}

type UpdateStepHandler struct {
	stepRepo     domainstep.StepWriteRepository
	connReadRepo domainconnection.ConnectionReadRepository
	variableRead domainvariable.VariableReadRepository
	outbox       port.OutboxRepository
}

func NewUpdateStepHandler(
	stepRepo domainstep.StepWriteRepository,
	connReadRepo domainconnection.ConnectionReadRepository,
	variableRead domainvariable.VariableReadRepository,
	outbox port.OutboxRepository,
) *UpdateStepHandler {
	return &UpdateStepHandler{
		stepRepo:     stepRepo,
		connReadRepo: connReadRepo,
		variableRead: variableRead,
		outbox:       outbox,
	}
}

func (h *UpdateStepHandler) Handle(
	ctx context.Context,
	cmd UpdateStepCommand,
) (*domainstep.Step, error) {
	if cmd.Name == "" {
		return nil, errors.New("name is required")
	}
	if cmd.URL == "" {
		return nil, errors.New("url is required")
	}
	if cmd.Method == "" {
		return nil, errors.New("method is required")
	}

	s, err := h.stepRepo.GetByID(ctx, cmd.ID)
	if err != nil {
		return nil, errors.New("failed to get step")
	}
	if s == nil || s.Status == domainstep.StatusDeleted {
		return nil, errors.New("step not found")
	}
	if s.OrganizationID != cmd.OrganizationID || s.WorkflowID != cmd.WorkflowID {
		return nil, errors.New("step not found")
	}

	if err := h.validateVariableReferences(ctx, cmd); err != nil {
		return nil, err
	}

	s.ApplyConfigUpdate(domainstep.UpdateStepConfigParams{
		Name:           cmd.Name,
		Description:    cmd.Description,
		URL:            cmd.URL,
		Method:         cmd.Method,
		Headers:        cmd.Headers,
		Query:          cmd.Query,
		Body:           cmd.Body,
		Timeout:        cmd.Timeout,
		RetryOnFailure: cmd.RetryOnFailure,
		RetryCount:     cmd.RetryCount,
		RetryDelay:     cmd.RetryDelay,
	})

	err = h.stepRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		if err := h.stepRepo.Update(txCtx, s); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, s.PullEvents())
	})
	if err != nil {
		return nil, errors.New("failed to update step")
	}

	return s, nil
}

func (h *UpdateStepHandler) validateVariableReferences(ctx context.Context, cmd UpdateStepCommand) error {
	refs := domainvariable.CollectReferencedIDs(cmd.URL, cmd.Headers, cmd.Query, cmd.Body)
	if len(refs) == 0 {
		return nil
	}

	views, err := h.variableRead.FindByIDs(ctx, refs)
	if err != nil {
		return errors.New("failed to validate variables")
	}
	byID := make(map[uuid.UUID]domainvariable.VariableView, len(views))
	for _, view := range views {
		byID[view.ID] = view
	}

	connections, err := h.connReadRepo.FindByWorkflowID(ctx, cmd.WorkflowID)
	if err != nil {
		return errors.New("failed to validate variables")
	}
	edges := make([]domainvariable.GraphEdge, 0, len(connections))
	for _, conn := range connections {
		edges = append(edges, domainvariable.GraphEdge{
			SourceStepID: conn.SourceStepID,
			TargetStepID: conn.TargetStepID,
		})
	}

	if err := domainvariable.ValidateReferences(cmd.ID, refs, byID, edges); err != nil {
		return err
	}
	return nil
}
