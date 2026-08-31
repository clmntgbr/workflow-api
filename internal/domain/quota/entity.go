package quota

import (
	"time"

	"github.com/google/uuid"
)

type Quota struct {
	ID   uuid.UUID
	Name string

	MaxProjectMembers  int
	MaxProjects             int
	MaxWorkflows            int
	MaxStepsPerWorkflow     int
	MaxEndpoints            int
	MaxVariablesPerWorkflow int
	MaxAssertionsPerWorkflow int

	MaxWorkflowRunsPerMonth    int
	MaxConcurrentRuns          int
	MinScheduleIntervalMinutes int

	RunHistoryRetentionDays int

	MaxStepTimeoutSeconds  int
	MaxRetryCountPerStep   int
	MaxRequestBodySizeKB   int
	MaxResponseBodySizeKB  int

	AllowsOpenAPIImport bool
	AllowsInsights      bool
	AllowsDataExport    bool
	ExecutorPriority    int

	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewQuotaParams struct {
	Name                       string
	MaxProjectMembers     int
	MaxProjects                int
	MaxWorkflows               int
	MaxStepsPerWorkflow        int
	MaxEndpoints               int
	MaxVariablesPerWorkflow    int
	MaxAssertionsPerWorkflow   int
	MaxWorkflowRunsPerMonth    int
	MaxConcurrentRuns          int
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

func NewQuota(p NewQuotaParams) *Quota {
	now := time.Now().UTC()
	return &Quota{
		ID:                         uuid.New(),
		Name:                       p.Name,
		MaxProjectMembers:     p.MaxProjectMembers,
		MaxProjects:                p.MaxProjects,
		MaxWorkflows:               p.MaxWorkflows,
		MaxStepsPerWorkflow:        p.MaxStepsPerWorkflow,
		MaxEndpoints:               p.MaxEndpoints,
		MaxVariablesPerWorkflow:    p.MaxVariablesPerWorkflow,
		MaxAssertionsPerWorkflow: p.MaxAssertionsPerWorkflow,
		MaxWorkflowRunsPerMonth:    p.MaxWorkflowRunsPerMonth,
		MaxConcurrentRuns:          p.MaxConcurrentRuns,
		MinScheduleIntervalMinutes: p.MinScheduleIntervalMinutes,
		RunHistoryRetentionDays:    p.RunHistoryRetentionDays,
		MaxStepTimeoutSeconds:      p.MaxStepTimeoutSeconds,
		MaxRetryCountPerStep:       p.MaxRetryCountPerStep,
		MaxRequestBodySizeKB:       p.MaxRequestBodySizeKB,
		MaxResponseBodySizeKB:      p.MaxResponseBodySizeKB,
		AllowsOpenAPIImport:        p.AllowsOpenAPIImport,
		AllowsInsights:             p.AllowsInsights,
		AllowsDataExport:           p.AllowsDataExport,
		ExecutorPriority:           p.ExecutorPriority,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}
}

func (q *Quota) ApplyUpdate(p NewQuotaParams) {
	q.Name = p.Name
	q.MaxProjectMembers = p.MaxProjectMembers
	q.MaxProjects = p.MaxProjects
	q.MaxWorkflows = p.MaxWorkflows
	q.MaxStepsPerWorkflow = p.MaxStepsPerWorkflow
	q.MaxEndpoints = p.MaxEndpoints
	q.MaxVariablesPerWorkflow = p.MaxVariablesPerWorkflow
	q.MaxAssertionsPerWorkflow = p.MaxAssertionsPerWorkflow
	q.MaxWorkflowRunsPerMonth = p.MaxWorkflowRunsPerMonth
	q.MaxConcurrentRuns = p.MaxConcurrentRuns
	q.MinScheduleIntervalMinutes = p.MinScheduleIntervalMinutes
	q.RunHistoryRetentionDays = p.RunHistoryRetentionDays
	q.MaxStepTimeoutSeconds = p.MaxStepTimeoutSeconds
	q.MaxRetryCountPerStep = p.MaxRetryCountPerStep
	q.MaxRequestBodySizeKB = p.MaxRequestBodySizeKB
	q.MaxResponseBodySizeKB = p.MaxResponseBodySizeKB
	q.AllowsOpenAPIImport = p.AllowsOpenAPIImport
	q.AllowsInsights = p.AllowsInsights
	q.AllowsDataExport = p.AllowsDataExport
	q.ExecutorPriority = p.ExecutorPriority
	q.UpdatedAt = time.Now().UTC()
}
