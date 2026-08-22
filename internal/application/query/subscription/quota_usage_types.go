package subscription

import "time"

type QuotaCounter struct {
	Used int64
	Max  int
	Left int64
}

type MonthlyQuotaCounter struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	Used        int64
	Max         int
	Left        int64
}

type QuotaLimits struct {
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

type QuotaUsageView struct {
	WorkflowRuns   MonthlyQuotaCounter
	Workflows      QuotaCounter
	Endpoints      QuotaCounter
	Members        QuotaCounter
	ConcurrentRuns QuotaCounter
	Limits         QuotaLimits
}
