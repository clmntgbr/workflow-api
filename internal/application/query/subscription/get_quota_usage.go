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

type QuotaUsageView struct {
	PeriodStart time.Time
	PeriodEnd   time.Time

	WorkflowRunsUsed int64
	WorkflowRunsMax  int
	WorkflowRunsLeft int64

	WorkflowsUsed int64
	WorkflowsMax  int
	WorkflowsLeft int64

	EndpointsUsed int64
	EndpointsMax  int
	EndpointsLeft int64

	MembersUsed int64
	MembersMax  int
	MembersLeft int64

	ConcurrentRunsUsed int64
	ConcurrentRunsMax  int
	ConcurrentRunsLeft int64

	MaxStepsPerWorkflow        int
	MaxVariablesPerWorkflow    int
	MinScheduleIntervalMinutes int
	RunHistoryRetentionDays    int
	MaxStepTimeoutSeconds      int
	MaxRetryCountPerStep       int
	MaxRequestBodySizeKB       int
	MaxResponseBodySizeKB      int
	AllowsOpenAPIImport        bool
	AllowsInsights             bool
	AllowsDataExport           bool
	ExecutorPriority           int
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
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,

		WorkflowRunsUsed: runStats.TotalRuns,
		WorkflowRunsMax:  quota.MaxWorkflowRunsPerMonth,
		WorkflowRunsLeft: quotaLeft(quota.MaxWorkflowRunsPerMonth, runStats.TotalRuns),

		WorkflowsUsed: workflowsTotal,
		WorkflowsMax:  quota.MaxWorkflows,
		WorkflowsLeft: quotaLeft(quota.MaxWorkflows, workflowsTotal),

		EndpointsUsed: endpointsTotal,
		EndpointsMax:  quota.MaxEndpoints,
		EndpointsLeft: quotaLeft(quota.MaxEndpoints, endpointsTotal),

		MembersUsed: membersUsed,
		MembersMax:  quota.MaxOrganizationMembers,
		MembersLeft: quotaLeft(quota.MaxOrganizationMembers, membersUsed),

		ConcurrentRunsUsed: concurrentUsed,
		ConcurrentRunsMax:  quota.MaxConcurrentRuns,
		ConcurrentRunsLeft: quotaLeft(quota.MaxConcurrentRuns, concurrentUsed),

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
	}, nil
}

func quotaLeft(max int, used int64) int64 {
	left := int64(max) - used
	if left < 0 {
		return 0
	}
	return left
}
