package presenter

import (
	"time"

	querysubscription "go-api/internal/application/query/subscription"
)

type QuotaUsageResponse struct {
	PeriodStart time.Time `json:"periodStart"`
	PeriodEnd   time.Time `json:"periodEnd"`

	WorkflowRunsUsed int64 `json:"workflowRunsUsed"`
	WorkflowRunsMax  int   `json:"workflowRunsMax"`
	WorkflowRunsLeft int64 `json:"workflowRunsLeft"`

	WorkflowsUsed int64 `json:"workflowsUsed"`
	WorkflowsMax  int   `json:"workflowsMax"`
	WorkflowsLeft int64 `json:"workflowsLeft"`

	EndpointsUsed int64 `json:"endpointsUsed"`
	EndpointsMax  int   `json:"endpointsMax"`
	EndpointsLeft int64 `json:"endpointsLeft"`

	MembersUsed int64 `json:"membersUsed"`
	MembersMax  int   `json:"membersMax"`
	MembersLeft int64 `json:"membersLeft"`

	ConcurrentRunsUsed int64 `json:"concurrentRunsUsed"`
	ConcurrentRunsMax  int   `json:"concurrentRunsMax"`
	ConcurrentRunsLeft int64 `json:"concurrentRunsLeft"`

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

func NewQuotaUsageResponse(usage *querysubscription.QuotaUsageView) QuotaUsageResponse {
	return QuotaUsageResponse{
		PeriodStart: usage.PeriodStart,
		PeriodEnd:   usage.PeriodEnd,

		WorkflowRunsUsed: usage.WorkflowRunsUsed,
		WorkflowRunsMax:  usage.WorkflowRunsMax,
		WorkflowRunsLeft: usage.WorkflowRunsLeft,

		WorkflowsUsed: usage.WorkflowsUsed,
		WorkflowsMax:  usage.WorkflowsMax,
		WorkflowsLeft: usage.WorkflowsLeft,

		EndpointsUsed: usage.EndpointsUsed,
		EndpointsMax:  usage.EndpointsMax,
		EndpointsLeft: usage.EndpointsLeft,

		MembersUsed: usage.MembersUsed,
		MembersMax:  usage.MembersMax,
		MembersLeft: usage.MembersLeft,

		ConcurrentRunsUsed: usage.ConcurrentRunsUsed,
		ConcurrentRunsMax:  usage.ConcurrentRunsMax,
		ConcurrentRunsLeft: usage.ConcurrentRunsLeft,

		MaxStepsPerWorkflow:        usage.MaxStepsPerWorkflow,
		MaxVariablesPerWorkflow:    usage.MaxVariablesPerWorkflow,
		MinScheduleIntervalMinutes: usage.MinScheduleIntervalMinutes,
		RunHistoryRetentionDays:    usage.RunHistoryRetentionDays,
		MaxStepTimeoutSeconds:      usage.MaxStepTimeoutSeconds,
		MaxRetryCountPerStep:       usage.MaxRetryCountPerStep,
		MaxRequestBodySizeKB:       usage.MaxRequestBodySizeKB,
		MaxResponseBodySizeKB:      usage.MaxResponseBodySizeKB,
		AllowsOpenAPIImport:        usage.AllowsOpenAPIImport,
		AllowsInsights:             usage.AllowsInsights,
		AllowsDataExport:           usage.AllowsDataExport,
		ExecutorPriority:           usage.ExecutorPriority,
	}
}
