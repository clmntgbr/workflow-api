package write

import (
	"time"

	domainquota "go-api/internal/domain/quota"

	"github.com/google/uuid"
)

type QuotaModel struct {
	ID   uuid.UUID `gorm:"column:id;primaryKey"`
	Name string    `gorm:"column:name"`

	MaxProjectMembers  int `gorm:"column:max_project_members"`
	MaxProjects             int `gorm:"column:max_projects"`
	MaxWorkflows            int `gorm:"column:max_workflows"`
	MaxStepsPerWorkflow     int `gorm:"column:max_steps_per_workflow"`
	MaxEndpoints            int `gorm:"column:max_endpoints"`
	MaxVariablesPerWorkflow int `gorm:"column:max_variables_per_workflow"`

	MaxWorkflowRunsPerMonth    int `gorm:"column:max_workflow_runs_per_month"`
	MaxConcurrentRuns          int `gorm:"column:max_concurrent_runs"`
	MinScheduleIntervalMinutes int `gorm:"column:min_schedule_interval_minutes"`

	RunHistoryRetentionDays int `gorm:"column:run_history_retention_days"`

	MaxStepTimeoutSeconds int `gorm:"column:max_step_timeout_seconds"`
	MaxRetryCountPerStep  int `gorm:"column:max_retry_count_per_step"`
	MaxRequestBodySizeKB  int `gorm:"column:max_request_body_size_kb"`
	MaxResponseBodySizeKB int `gorm:"column:max_response_body_size_kb"`

	AllowsOpenAPIImport bool `gorm:"column:allows_openapi_import"`
	AllowsInsights      bool `gorm:"column:allows_insights"`
	AllowsDataExport    bool `gorm:"column:allows_data_export"`
	ExecutorPriority    int  `gorm:"column:executor_priority"`

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (QuotaModel) TableName() string { return "quotas" }

func quotaModelFromDomain(q *domainquota.Quota) *QuotaModel {
	return &QuotaModel{
		ID:                         q.ID,
		Name:                       q.Name,
		MaxProjectMembers:     q.MaxProjectMembers,
		MaxProjects:                q.MaxProjects,
		MaxWorkflows:               q.MaxWorkflows,
		MaxStepsPerWorkflow:        q.MaxStepsPerWorkflow,
		MaxEndpoints:               q.MaxEndpoints,
		MaxVariablesPerWorkflow:    q.MaxVariablesPerWorkflow,
		MaxWorkflowRunsPerMonth:    q.MaxWorkflowRunsPerMonth,
		MaxConcurrentRuns:          q.MaxConcurrentRuns,
		MinScheduleIntervalMinutes: q.MinScheduleIntervalMinutes,
		RunHistoryRetentionDays:    q.RunHistoryRetentionDays,
		MaxStepTimeoutSeconds:      q.MaxStepTimeoutSeconds,
		MaxRetryCountPerStep:       q.MaxRetryCountPerStep,
		MaxRequestBodySizeKB:       q.MaxRequestBodySizeKB,
		MaxResponseBodySizeKB:      q.MaxResponseBodySizeKB,
		AllowsOpenAPIImport:        q.AllowsOpenAPIImport,
		AllowsInsights:             q.AllowsInsights,
		AllowsDataExport:           q.AllowsDataExport,
		ExecutorPriority:           q.ExecutorPriority,
		CreatedAt:                  q.CreatedAt,
		UpdatedAt:                  q.UpdatedAt,
	}
}

func quotaDomainFromModel(m *QuotaModel) *domainquota.Quota {
	return &domainquota.Quota{
		ID:                         m.ID,
		Name:                       m.Name,
		MaxProjectMembers:     m.MaxProjectMembers,
		MaxProjects:                m.MaxProjects,
		MaxWorkflows:               m.MaxWorkflows,
		MaxStepsPerWorkflow:        m.MaxStepsPerWorkflow,
		MaxEndpoints:               m.MaxEndpoints,
		MaxVariablesPerWorkflow:    m.MaxVariablesPerWorkflow,
		MaxWorkflowRunsPerMonth:    m.MaxWorkflowRunsPerMonth,
		MaxConcurrentRuns:          m.MaxConcurrentRuns,
		MinScheduleIntervalMinutes: m.MinScheduleIntervalMinutes,
		RunHistoryRetentionDays:    m.RunHistoryRetentionDays,
		MaxStepTimeoutSeconds:      m.MaxStepTimeoutSeconds,
		MaxRetryCountPerStep:       m.MaxRetryCountPerStep,
		MaxRequestBodySizeKB:       m.MaxRequestBodySizeKB,
		MaxResponseBodySizeKB:      m.MaxResponseBodySizeKB,
		AllowsOpenAPIImport:        m.AllowsOpenAPIImport,
		AllowsInsights:             m.AllowsInsights,
		AllowsDataExport:           m.AllowsDataExport,
		ExecutorPriority:           m.ExecutorPriority,
		CreatedAt:                  m.CreatedAt,
		UpdatedAt:                  m.UpdatedAt,
	}
}

func quotaViewFromModel(m *QuotaModel) domainquota.QuotaView {
	return domainquota.QuotaView{
		ID:                         m.ID,
		Name:                       m.Name,
		MaxProjectMembers:     m.MaxProjectMembers,
		MaxProjects:                m.MaxProjects,
		MaxWorkflows:               m.MaxWorkflows,
		MaxStepsPerWorkflow:        m.MaxStepsPerWorkflow,
		MaxEndpoints:               m.MaxEndpoints,
		MaxVariablesPerWorkflow:    m.MaxVariablesPerWorkflow,
		MaxWorkflowRunsPerMonth:    m.MaxWorkflowRunsPerMonth,
		MaxConcurrentRuns:          m.MaxConcurrentRuns,
		MinScheduleIntervalMinutes: m.MinScheduleIntervalMinutes,
		RunHistoryRetentionDays:    m.RunHistoryRetentionDays,
		MaxStepTimeoutSeconds:      m.MaxStepTimeoutSeconds,
		MaxRetryCountPerStep:       m.MaxRetryCountPerStep,
		MaxRequestBodySizeKB:       m.MaxRequestBodySizeKB,
		MaxResponseBodySizeKB:      m.MaxResponseBodySizeKB,
		AllowsOpenAPIImport:        m.AllowsOpenAPIImport,
		AllowsInsights:             m.AllowsInsights,
		AllowsDataExport:           m.AllowsDataExport,
		ExecutorPriority:           m.ExecutorPriority,
		CreatedAt:                  m.CreatedAt,
		UpdatedAt:                  m.UpdatedAt,
	}
}
