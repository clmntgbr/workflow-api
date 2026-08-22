package subscription

import (
	"context"
	"errors"
	"time"

	domainendpoint "go-api/internal/domain/endpoint"
	domainproject "go-api/internal/domain/project"
	"go-api/internal/domain/paginate"
	domainsubscription "go-api/internal/domain/subscription"
	domainuser "go-api/internal/domain/user"
	domainworkflow "go-api/internal/domain/workflow"
	domainworkflowrun "go-api/internal/domain/workflowrun"

	"github.com/google/uuid"
)

var ErrActiveProjectRequired = errors.New("active project is required")

type GetQuotaUsageQuery struct {
	UserID         uuid.UUID
	ProjectID uuid.UUID
}

type GetQuotaUsageHandler struct {
	userRepo         domainuser.UserReadRepository
	subscriptionRepo domainsubscription.SubscriptionReadRepository
	projectRepo      domainproject.ProjectReadRepository
	workflowRepo     domainworkflow.WorkflowReadRepository
	endpointRepo     domainendpoint.EndpointReadRepository
	workflowRunRepo  domainworkflowrun.WorkflowRunReadRepository
}

func NewGetQuotaUsageHandler(
	userRepo domainuser.UserReadRepository,
	subscriptionRepo domainsubscription.SubscriptionReadRepository,
	projectRepo domainproject.ProjectReadRepository,
	workflowRepo domainworkflow.WorkflowReadRepository,
	endpointRepo domainendpoint.EndpointReadRepository,
	workflowRunRepo domainworkflowrun.WorkflowRunReadRepository,
) *GetQuotaUsageHandler {
	return &GetQuotaUsageHandler{
		userRepo:         userRepo,
		subscriptionRepo: subscriptionRepo,
		projectRepo: projectRepo,
		workflowRepo:     workflowRepo,
		endpointRepo:     endpointRepo,
		workflowRunRepo:  workflowRunRepo,
	}
}

func (h *GetQuotaUsageHandler) Handle(ctx context.Context, q GetQuotaUsageQuery) (*QuotaUsageView, error) {
	if q.ProjectID == uuid.Nil {
		return nil, ErrActiveProjectRequired
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

	org, err := h.projectRepo.FindByID(ctx, q.ProjectID)
	if err != nil {
		return nil, errors.New("failed to get project")
	}
	if org == nil {
		return nil, errors.New("project not found")
	}

	anchor := sub.QuotaPeriodStart
	if anchor.IsZero() {
		anchor = sub.StartDate
	}
	if anchor.IsZero() {
		anchor = sub.CreatedAt
	}

	periodStart, periodEnd := domainsubscription.CurrentQuotaPeriod(anchor, time.Now().UTC())

	runStats, err := h.workflowRunRepo.FindAnalyticsByProject(ctx, q.ProjectID, domainworkflowrun.WorkflowRunAnalyticsFilter{
		From: &periodStart,
		To:   &periodEnd,
	})
	if err != nil {
		return nil, errors.New("failed to count workflow run usage")
	}

	concurrentStats, err := h.workflowRunRepo.FindAnalyticsByProject(ctx, q.ProjectID, domainworkflowrun.WorkflowRunAnalyticsFilter{})
	if err != nil {
		return nil, errors.New("failed to count concurrent runs")
	}

	projects, err := h.projectRepo.FindByUserID(ctx, q.UserID)
	if err != nil {
		return nil, errors.New("failed to count projects")
	}
	projectsTotal := int64(len(projects))

	countQuery := paginate.PaginateQuery{Page: 1, Limit: 1}
	countQuery.Normalize()

	_, workflowsTotal, err := h.workflowRepo.FindByProjectID(ctx, q.ProjectID, countQuery)
	if err != nil {
		return nil, errors.New("failed to count workflows")
	}

	_, endpointsTotal, err := h.endpointRepo.FindByProjectID(ctx, q.ProjectID, domainendpoint.ListEndpointsFilter{
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
		Projects: QuotaCounter{
			Used: projectsTotal,
			Max:  quota.MaxProjects,
			Left: quotaLeft(quota.MaxProjects, projectsTotal),
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
			Max:  quota.MaxProjectMembers,
			Left: quotaLeft(quota.MaxProjectMembers, membersUsed),
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
