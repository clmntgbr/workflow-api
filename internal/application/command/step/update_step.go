package step

import (
	"context"
	"errors"

	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

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
	stepRepo domainstep.StepWriteRepository
	outbox   port.OutboxRepository
}

func NewUpdateStepHandler(
	stepRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
) *UpdateStepHandler {
	return &UpdateStepHandler{
		stepRepo: stepRepo,
		outbox:   outbox,
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
