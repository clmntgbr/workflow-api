package presenter

import (
	"time"

	querysubscription "go-api/internal/application/query/subscription"
)

type MonthlyQuotaCounterResponse struct {
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`
	Used        int64     `json:"used"`
	Max         int       `json:"max"`
	Left        int64     `json:"left"`
}

type QuotaCounterResponse struct {
	Used int64 `json:"used"`
	Max  int   `json:"max"`
	Left int64 `json:"left"`
}

type QuotaLimitsResponse struct {
	MaxStepsPerWorkflow        int  `json:"maxStepsPerWorkflow"`
	MaxVariablesPerWorkflow    int  `json:"maxVariablesPerWorkflow"`
	MinScheduleIntervalMinutes int  `json:"minScheduleIntervalMinutes"`
	RunHistoryRetentionDays    int  `json:"runHistoryRetentionDays"`
	MaxStepTimeoutSeconds      int  `json:"maxStepTimeoutSeconds"`
	MaxRetryCountPerStep       int  `json:"maxRetryCountPerStep"`
	MaxRequestBodySizeKB       int  `json:"maxRequestBodySizeKb"`
	MaxResponseBodySizeKB      int  `json:"maxResponseBodySizeKb"`
	AllowsOpenAPIImport        bool `json:"allowsOpenApiImport"`
	AllowsInsights             bool `json:"allowsInsights"`
	AllowsDataExport           bool `json:"allowsDataExport"`
	ExecutorPriority           int  `json:"executorPriority"`
}

type QuotaUsageResponse struct {
	WorkflowRuns   MonthlyQuotaCounterResponse `json:"workflowRuns"`
	Workflows      QuotaCounterResponse        `json:"workflows"`
	Endpoints      QuotaCounterResponse        `json:"endpoints"`
	Members        QuotaCounterResponse        `json:"members"`
	ConcurrentRuns QuotaCounterResponse        `json:"concurrentRuns"`
	Limits         QuotaLimitsResponse         `json:"limits"`
}

func NewQuotaUsageResponse(usage *querysubscription.QuotaUsageView) QuotaUsageResponse {
	return QuotaUsageResponse{
		WorkflowRuns: MonthlyQuotaCounterResponse{
			PeriodStart: usage.WorkflowRuns.PeriodStart,
			PeriodEnd:   usage.WorkflowRuns.PeriodEnd,
			Used:        usage.WorkflowRuns.Used,
			Max:         usage.WorkflowRuns.Max,
			Left:        usage.WorkflowRuns.Left,
		},
		Workflows: QuotaCounterResponse{
			Used: usage.Workflows.Used,
			Max:  usage.Workflows.Max,
			Left: usage.Workflows.Left,
		},
		Endpoints: QuotaCounterResponse{
			Used: usage.Endpoints.Used,
			Max:  usage.Endpoints.Max,
			Left: usage.Endpoints.Left,
		},
		Members: QuotaCounterResponse{
			Used: usage.Members.Used,
			Max:  usage.Members.Max,
			Left: usage.Members.Left,
		},
		ConcurrentRuns: QuotaCounterResponse{
			Used: usage.ConcurrentRuns.Used,
			Max:  usage.ConcurrentRuns.Max,
			Left: usage.ConcurrentRuns.Left,
		},
		Limits: QuotaLimitsResponse{
			MaxStepsPerWorkflow:        usage.Limits.MaxStepsPerWorkflow,
			MaxVariablesPerWorkflow:    usage.Limits.MaxVariablesPerWorkflow,
			MinScheduleIntervalMinutes: usage.Limits.MinScheduleIntervalMinutes,
			RunHistoryRetentionDays:    usage.Limits.RunHistoryRetentionDays,
			MaxStepTimeoutSeconds:      usage.Limits.MaxStepTimeoutSeconds,
			MaxRetryCountPerStep:       usage.Limits.MaxRetryCountPerStep,
			MaxRequestBodySizeKB:       usage.Limits.MaxRequestBodySizeKB,
			MaxResponseBodySizeKB:      usage.Limits.MaxResponseBodySizeKB,
			AllowsOpenAPIImport:        usage.Limits.AllowsOpenAPIImport,
			AllowsInsights:             usage.Limits.AllowsInsights,
			AllowsDataExport:           usage.Limits.AllowsDataExport,
			ExecutorPriority:           usage.Limits.ExecutorPriority,
		},
	}
}
