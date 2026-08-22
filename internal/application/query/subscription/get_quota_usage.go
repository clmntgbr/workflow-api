package subscription

import (
	"context"
	"errors"
	"time"

	domainendpoint "go-api/internal/domain/endpoint"
	domainorganization "go-api/internal/domain/organization"
	"go-api/internal/domain/paginate"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

var ErrActiveOrganizationRequired = errors.New("active organization is required")

type GetQuotaUsageQuery struct {
	UserID         uuid.UUID
	OrganizationID uuid.UUID
}

type GetQuotaUsageHandler struct {
	userRepo         domainuser.UserReadRepository
	subscriptionRepo domainsubscription.SubscriptionReadRepository
	orgRepo          domainorganization.OrganizationReadRepository
	workflowRepo     domainworkflow.WorkflowReadRepository
	endpointRepo     domainendpoint.EndpointReadRepository
	workflowRunRepo  domainworkflowrun.WorkflowRunReadRepository
}

func NewGetQuotaUsageHandler(
	userRepo domainuser.UserReadRepository,
	subscriptionRepo domainsubscription.SubscriptionReadRepository,
	orgRepo domainorganization.OrganizationReadRepository,
	workflowRepo domainworkflow.WorkflowReadRepository,
	endpointRepo domainendpoint.EndpointReadRepository,
	workflowRunRepo domainworkflowrun.WorkflowRunReadRepository,
) *GetQuotaUsageHandler {
	return &GetQuotaUsageHandler{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		orgRepo:          orgRepo,
		workflowRepo:     workflowRepo,
		endpointRepo:     endpointRepo,
		workflowRunRepo:  workflowRunRepo,
	}
}

func (h *GetQuotaUsageHandler) Handle(ctx context.Context, q GetQuotaUsageQuery) (*QuotaUsageView, error) {
	if q.OrganizationID == uuid.Nil {
		return nil, ErrActiveOrganizationRequired
	}

	user, err := h.userRepo.FindByID(ctx, q.UserID)
	if err != nil {
		return nil, errors.New("failed to get user")
	}
	if user == nil || user.SubscriptionID == nil {
		return nil, ErrSubscriptionNotFound
	}

	sub, err := h.subscriptionRepo.FindByID(ctx, *user.SubscriptionID)
	if err != nil {
		return nil, errors.New("failed to get subscription")
	}
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	if sub.Plan == nil || sub.Plan.Quota == nil {
		return nil, errors.New("subscription plan quota not found")
	}

	org, err := h.orgRepo.FindByID(ctx, q.OrganizationID)
	if err != nil {
		return nil, errors.New("failed to get organization")
	}
	if org == nil {
		return nil, errors.New("organization not found")
	}

	anchor := sub.QuotaPeriodStart
	if anchor.IsZero() {
		anchor = sub.StartDate
	}
	if anchor.IsZero() {
		anchor = sub.CreatedAt
	}

	periodStart, periodEnd := domainsubscription.CurrentQuotaPeriod(anchor, time.Now().UTC())

	runStats, err := h.workflowRunRepo.FindAnalyticsByOrganization(ctx, q.OrganizationID, domainworkflowrun.WorkflowRunAnalyticsFilter{
		From: &periodStart,
		To:   &periodEnd,
	})
	if err != nil {
		return nil, errors.New("failed to count workflow run usage")
	}

	concurrentStats, err := h.workflowRunRepo.FindAnalyticsByOrganization(ctx, q.OrganizationID, domainworkflowrun.WorkflowRunAnalyticsFilter{})
	if err != nil {
		return nil, errors.New("failed to count concurrent runs")
	}

	countQuery := paginate.PaginateQuery{Page: 1, Limit: 1}
	countQuery.Normalize()

	_, workflowsTotal, err := h.workflowRepo.FindByOrganizationID(ctx, q.OrganizationID, countQuery)
	if err != nil {
		return nil, errors.New("failed to count workflows")
	}

	_, endpointsTotal, err := h.endpointRepo.FindByOrganizationID(ctx, q.OrganizationID, domainendpoint.ListEndpointsFilter{
		PaginateQuery: countQuery,
	})
	if err != nil {
		return nil, errors.New("failed to count endpoints")
	}

	quota := sub.Plan.Quota
	membersUsed := int64(len(org.MemberIDs))
	concurrentUsed := concurrentStats.RunningCount + concurrentStats.PendingCount

	return &QuotaUsageView{
		WorkflowRuns: MonthlyQuotaCounter{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Used:        runStats.TotalRuns,
			Max:         quota.MaxWorkflowRunsPerMonth,
			Left:        quotaLeft(quota.MaxWorkflowRunsPerMonth, runStats.TotalRuns),
		},
		Workflows: QuotaCounter{
			Used: workflowsTotal,
			Max:  quota.MaxWorkflows,
			Left: quotaLeft(quota.MaxWorkflows, workflowsTotal),
		},
		Endpoints: QuotaCounter{
			Used: endpointsTotal,
			Max:  quota.MaxEndpoints,
			Left: quotaLeft(quota.MaxEndpoints, endpointsTotal),
		},
		Members: QuotaCounter{
			Used: membersUsed,
			Max:  quota.MaxOrganizationMembers,
			Left: quotaLeft(quota.MaxOrganizationMembers, membersUsed),
		},
		ConcurrentRuns: QuotaCounter{
			Used: concurrentUsed,
			Max:  quota.MaxConcurrentRuns,
			Left: quotaLeft(quota.MaxConcurrentRuns, concurrentUsed),
		},
		Limits: QuotaLimits{
			MaxStepsPerWorkflow:        quota.MaxStepsPerWorkflow,
			MaxVariablesPerWorkflow:    quota.MaxVariablesPerWorkflow,
			MinScheduleIntervalMinutes: quota.MinScheduleIntervalMinutes,
			RunHistoryRetentionDays:    quota.RunHistoryRetentionDays,
			MaxStepTimeoutSeconds:      quota.MaxStepTimeoutSeconds,
			MaxRetryCountPerStep:       quota.MaxRetryCountPerStep,
			MaxRequestBodySizeKB:       quota.MaxRequestBodySizeKB,
			MaxResponseBodySizeKB:      quota.MaxResponseBodySizeKB,
			AllowsOpenAPIImport:        quota.AllowsOpenAPIImport,
			AllowsInsights:             quota.AllowsInsights,
			AllowsDataExport:           quota.AllowsDataExport,
			ExecutorPriority:           quota.ExecutorPriority,
		},
	}, nil
}

func quotaLeft(max int, used int64) int64 {
	left := int64(max) - used
	if left < 0 {
		return 0
	}
	return left
}
