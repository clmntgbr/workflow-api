package quota

import (
	"context"
	"errors"

	querysubscription "go-api/internal/application/query/subscription"
	domainassertion "go-api/internal/domain/assertion"
	domainproject "go-api/internal/domain/project"
	domainstep "go-api/internal/domain/step"
	domainuser "go-api/internal/domain/user"
	domainvariable "go-api/internal/domain/variable"

	"github.com/google/uuid"
)

var (
	ErrWorkflowQuotaExceeded      = errors.New("workflow quota exceeded for your current plan")
	ErrEndpointQuotaExceeded      = errors.New("endpoint quota exceeded for your current plan")
	ErrStepQuotaExceeded          = errors.New("step quota exceeded for your current plan")
	ErrVariableQuotaExceeded      = errors.New("variable quota exceeded for your current plan")
	ErrAssertionQuotaExceeded     = errors.New("assertion quota exceeded for your current plan")
	ErrWorkflowRunQuotaExceeded   = errors.New("workflow run quota exceeded for your current plan")
	ErrConcurrentRunQuotaExceeded = errors.New("concurrent run quota exceeded for your current plan")
	ErrProjectQuotaExceeded       = errors.New("project quota exceeded for your current plan")
)

type AssertCreateAllowedHandler struct {
	getQuotaUsage     *querysubscription.GetQuotaUsageHandler
	stepReadRepo      domainstep.StepReadRepository
	variableReadRepo  domainvariable.VariableReadRepository
	assertionReadRepo domainassertion.AssertionReadRepository
	projectReadRepo   domainproject.ProjectReadRepository
	userReadRepo      domainuser.UserReadRepository
}

func NewAssertCreateAllowedHandler(
	getQuotaUsage *querysubscription.GetQuotaUsageHandler,
	stepReadRepo domainstep.StepReadRepository,
	variableReadRepo domainvariable.VariableReadRepository,
	assertionReadRepo domainassertion.AssertionReadRepository,
	projectReadRepo domainproject.ProjectReadRepository,
	userReadRepo domainuser.UserReadRepository,
) *AssertCreateAllowedHandler {
	return &AssertCreateAllowedHandler{
		getQuotaUsage:     getQuotaUsage,
		stepReadRepo:      stepReadRepo,
		variableReadRepo:  variableReadRepo,
		assertionReadRepo: assertionReadRepo,
		projectReadRepo:   projectReadRepo,
		userReadRepo:      userReadRepo,
	}
}

func (h *AssertCreateAllowedHandler) AssertProjectCreate(
	ctx context.Context,
	userID uuid.UUID,
) error {
	projects, err := h.projectReadRepo.FindByUserID(ctx, userID)
	if err != nil {
		return errors.New("failed to count projects")
	}
	if len(projects) == 0 {
		return nil
	}

	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:    userID,
		ProjectID: projects[0].ID,
	})
	if err != nil {
		return err
	}
	if usage.Projects.Left <= 0 {
		return ErrProjectQuotaExceeded
	}
	return nil
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

func (h *AssertCreateAllowedHandler) AssertVariableCreate(
	ctx context.Context,
	userID uuid.UUID,
	projectID uuid.UUID,
	workflowID uuid.UUID,
) error {
	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}

	variables, err := h.variableReadRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return errors.New("failed to count workflow variables")
	}
	if int64(len(variables)) >= int64(usage.Limits.MaxVariablesPerWorkflow) {
		return ErrVariableQuotaExceeded
	}
	return nil
}

func (h *AssertCreateAllowedHandler) AssertAssertionCreate(
	ctx context.Context,
	userID uuid.UUID,
	projectID uuid.UUID,
	workflowID uuid.UUID,
) error {
	usage, err := h.getQuotaUsage.Handle(ctx, querysubscription.GetQuotaUsageQuery{
		UserID:    userID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}

	assertions, err := h.assertionReadRepo.FindByWorkflowID(ctx, workflowID)
	if err != nil {
		return errors.New("failed to count workflow assertions")
	}
	if int64(len(assertions)) >= int64(usage.Limits.MaxAssertionsPerWorkflow) {
		return ErrAssertionQuotaExceeded
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
