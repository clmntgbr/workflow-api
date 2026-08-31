package step

import (
	"context"
	"errors"

	cmdquota "go-api/internal/application/command/quota"
	"go-api/internal/application/messaging"
	"go-api/internal/domain/httpquery"
	"go-api/internal/domain/port"
	domainstep "go-api/internal/domain/step"

	"github.com/google/uuid"
)

type UpdateStepCommand struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	WorkflowID     uuid.UUID
	ProjectID      uuid.UUID
	Name           string
	Description    string
	URL            string
	Method         string
	Headers        map[string]string
	Query          httpquery.Params
	Body           map[string]any
	Timeout        int
	RetryOnFailure bool
	RetryCount     int
	RetryDelay     int
}

type UpdateStepHandler struct {
	stepRepo domainstep.StepWriteRepository
	outbox   port.OutboxRepository
	assert   *cmdquota.AssertCreateAllowedHandler
}

func NewUpdateStepHandler(
	stepRepo domainstep.StepWriteRepository,
	outbox port.OutboxRepository,
	assert *cmdquota.AssertCreateAllowedHandler,
) *UpdateStepHandler {
	return &UpdateStepHandler{
		stepRepo: stepRepo,
		outbox:   outbox,
		assert:   assert,
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
	if s.ProjectID != cmd.ProjectID || s.WorkflowID != cmd.WorkflowID {
		return nil, errors.New("step not found")
	}
	if s.Type != domainstep.TypeHTTP {
		return nil, errors.New("only HTTP steps can be updated with HTTP configuration")
	}

	if err := h.assert.AssertStepHTTPConfig(ctx, cmd.UserID, cmd.ProjectID, cmd.Timeout, cmd.RetryCount); err != nil {
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
		if err := domainstep.ValidateConfig(s); err != nil {
			return err
		}
		if err := h.stepRepo.Update(txCtx, s); err != nil {
			return err
		}
		return h.outbox.StoreEvents(txCtx, messaging.WithPerformedBy(s.PullEvents(), cmd.UserID))
	})
	if err != nil {
		return nil, errors.New("failed to update step")
	}

	return s, nil
}
