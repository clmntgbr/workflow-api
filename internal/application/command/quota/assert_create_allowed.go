package quota

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	domainorganization "go-api/internal/domain/organization"
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
	orgReadRepo   domainorganization.OrganizationReadRepository
	userReadRepo  domainuser.UserReadRepository
}

func NewAssertCreateAllowedHandler(
	getQuotaUsage *querysubscription.GetQuotaUsageHandler,
	stepReadRepo domainstep.StepReadRepository,
	orgReadRepo domainorganization.OrganizationReadRepository,
	userReadRepo domainuser.UserReadRepository,
) *AssertCreateAllowedHandler {
	return &AssertCreateAllowedHandler{
		getQuotaUsage: getQuotaUsage,
		stepReadRepo:  stepReadRepo,
		orgReadRepo:   orgReadRepo,
		userReadRepo:  userReadRepo,
	}
}

func (h *AssertCreateAllowedHandler) AssertWorkflowCreate(
	ctx context.Context,
	userID uuid.UUID,
	organizationID uuid.UUID,
) error {
	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		OrganizationID: organizationID,
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
	organizationID uuid.UUID,
	count int,
) error {
	if count <= 0 {
		count = 1
	}

	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		OrganizationID: organizationID,
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
	organizationID uuid.UUID,
	workflowID uuid.UUID,
) error {
	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		OrganizationID: organizationID,
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
	organizationID uuid.UUID,
	preferredUserID *uuid.UUID,
	count int,
) error {
	if count <= 0 {
		count = 1
	}

	userID, err := h.resolveBillingUserID(ctx, organizationID, preferredUserID)
	if err != nil {
		return err
	}

	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:         userID,
		OrganizationID: organizationID,
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
	organizationID uuid.UUID,
	preferredUserID *uuid.UUID,
) (uuid.UUID, error) {
	if preferredUserID != nil && *preferredUserID != uuid.Nil {
		return *preferredUserID, nil
	}

	org, err := h.orgReadRepo.FindByID(ctx, organizationID)
	if err != nil {
		return uuid.Nil, errors.New("failed to resolve billing user")
	}
	if org == nil {
		return uuid.Nil, errors.New("organization not found")
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
