package quota

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	domainproject "go-api/internal/domain/project"
	domainstep "go-api/internal/domain/step"
	domainuser "go-api/internal/domain/user"

	"github.com/google/uuid"
)

var (
	ErrWorkflowQuotaExceeded    = errors.New("workflow quota exceeded for your current plan")
	ErrEndpointQuotaExceeded    = errors.New("endpoint quota exceeded for your current plan")
	ErrStepQuotaExceeded        = errors.New("step quota exceeded for your current plan")
	ErrWorkflowRunQuotaExceeded = errors.New("workflow run quota exceeded for your current plan")
	ErrConcurrentRunQuotaExceeded = errors.New("concurrent run quota exceeded for your current plan")
)

type AssertCreateAllowedHandler struct {
	getQuotaUsage *querysubscription.GetQuotaUsageHandler
	stepReadRepo  domainstep.StepReadRepository
	projectReadRepo   domainproject.ProjectReadRepository
	userReadRepo  domainuser.UserReadRepository
}

func NewAssertCreateAllowedHandler(
	getQuotaUsage *querysubscription.GetQuotaUsageHandler,
	stepReadRepo domainstep.StepReadRepository,
	projectReadRepo domainproject.ProjectReadRepository,
	userReadRepo domainuser.UserReadRepository,
) *AssertCreateAllowedHandler {
	return &AssertCreateAllowedHandler{
		getQuotaUsage: getQuotaUsage,
		stepReadRepo:  stepReadRepo,
		projectReadRepo:   projectReadRepo,
		userReadRepo:  userReadRepo,
	}
}

func (h *AssertCreateAllowedHandler) AssertWorkflowCreate(
	ctx context.Context,
	userID uuid.UUID,
	projectID uuid.UUID,
) error {
	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}
	if usage.Workflows.Left <= 0 {
		return ErrWorkflowQuotaExceeded
	}
	return nil
}

func (h *AssertCreateAllowedHandler) AssertEndpointCreate(
	ctx context.Context,
	userID uuid.UUID,
	projectID uuid.UUID,
	count int,
) error {
	if count <= 0 {
		count = 1
	}

	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}
	if usage.Endpoints.Left < int64(count) {
		return ErrEndpointQuotaExceeded
	}
	return nil
}

func (h *AssertCreateAllowedHandler) AssertStepCreate(
	ctx context.Context,
	userID uuid.UUID,
	projectID uuid.UUID,
	workflowID uuid.UUID,
) error {
	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}

	steps, err := h.stepReadRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return errors.New("failed to count workflow steps")
	}
	if int64(len(steps)) >= int64(usage.Limits.MaxStepsPerWorkflow) {
		return ErrStepQuotaExceeded
	}
	return nil
}

func (h *AssertCreateAllowedHandler) AssertWorkflowRunStart(
	ctx context.Context,
	projectID uuid.UUID,
	preferredUserID *uuid.UUID,
	count int,
) error {
	if count <= 0 {
		count = 1
	}

	userID, err := h.resolveBillingUserID(ctx, projectID, preferredUserID)
	if err != nil {
		return err
	}

	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}
	if usage.WorkflowRuns.Left < int64(count) {
		return ErrWorkflowRunQuotaExceeded
	}
	if usage.ConcurrentRuns.Left < int64(count) {
		return ErrConcurrentRunQuotaExceeded
	}
	return nil
}

func (h *AssertCreateAllowedHandler) resolveBillingUserID(
	ctx context.Context,
	projectID uuid.UUID,
	preferredUserID *uuid.UUID,
) (uuid.UUID, error) {
	if preferredUserID != nil && *preferredUserID != uuid.Nil {
		return *preferredUserID, nil
	}

	org, err := h.projectReadRepo.FindByID(ctx, projectID)
	if err != nil {
		return uuid.Nil, errors.New("failed to resolve billing user")
	}
	if org == nil {
		return uuid.Nil, errors.New("project not found")
	}

	for _, memberID := range org.MemberIDs {
		user, err := h.userReadRepo.FindByID(ctx, memberID)
		if err != nil {
			return uuid.Nil, errors.New("failed to resolve billing user")
		}
		if user != nil && user.SubscriptionID != nil {
			return user.ID, nil
		}
	}

	return uuid.Nil, querysubscription.ErrSubscriptionNotFound
}
