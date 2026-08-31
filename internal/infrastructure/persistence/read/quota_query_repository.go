package read

import (
	"context"
	"errors"
	"time"

	domainquota "go-api/internal/domain/quota"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type quotaRow struct {
	ID                         uuid.UUID `gorm:"column:id"`
	Name                       string    `gorm:"column:name"`
	MaxProjectMembers          int       `gorm:"column:max_project_members"`
	MaxProjects                int       `gorm:"column:max_projects"`
	MaxWorkflows               int       `gorm:"column:max_workflows"`
	MaxStepsPerWorkflow        int       `gorm:"column:max_steps_per_workflow"`
	MaxEndpoints               int       `gorm:"column:max_endpoints"`
	MaxVariablesPerWorkflow    int       `gorm:"column:max_variables_per_workflow"`
	MaxAssertionsPerWorkflow   int       `gorm:"column:max_assertions_per_workflow"`
	MaxWorkflowRunsPerMonth    int       `gorm:"column:max_workflow_runs_per_month"`
	MaxConcurrentRuns          int       `gorm:"column:max_concurrent_runs"`
	MinScheduleIntervalMinutes int       `gorm:"column:min_schedule_interval_minutes"`
	RunHistoryRetentionDays    int       `gorm:"column:run_history_retention_days"`
	MaxStepTimeoutSeconds      int       `gorm:"column:max_step_timeout_seconds"`
	MaxRetryCountPerStep       int       `gorm:"column:max_retry_count_per_step"`
	MaxRequestBodySizeKB       int       `gorm:"column:max_request_body_size_kb"`
	MaxResponseBodySizeKB      int       `gorm:"column:max_response_body_size_kb"`
	AllowsOpenAPIImport        bool      `gorm:"column:allows_openapi_import"`
	AllowsInsights             bool      `gorm:"column:allows_insights"`
	AllowsDataExport           bool      `gorm:"column:allows_data_export"`
	ExecutorPriority           int       `gorm:"column:executor_priority"`
	CreatedAt                  time.Time `gorm:"column:created_at"`
	UpdatedAt                  time.Time `gorm:"column:updated_at"`
}

func (quotaRow) TableName() string { return "quotas" }

type quotaReadRepository struct {
	db *gorm.DB
}

func NewQuotaReadRepository(db *gorm.DB) domainquota.QuotaReadRepository {
	return &quotaReadRepository{db: db}
}

func (r *quotaReadRepository) FindByID(ctx context.Context, id uuid.UUID) (*domainquota.QuotaView, error) {
	var row quotaRow
	err := r.db.WithContext(ctx).First(&row, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	view := toQuotaView(row)
	return &view, nil
}

func (r *quotaReadRepository) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domainquota.QuotaView, error) {
	if len(ids) == 0 {
		return []domainquota.QuotaView{}, nil
	}
	var rows []quotaRow
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domainquota.QuotaView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toQuotaView(row))
	}
	return out, nil
}

func toQuotaView(row quotaRow) domainquota.QuotaView {
	return domainquota.QuotaView{
		ID:                         row.ID,
		Name:                       row.Name,
		MaxProjectMembers:     row.MaxProjectMembers,
		MaxProjects:                row.MaxProjects,
		MaxWorkflows:               row.MaxWorkflows,
		MaxStepsPerWorkflow:        row.MaxStepsPerWorkflow,
		MaxEndpoints:               row.MaxEndpoints,
		MaxVariablesPerWorkflow:    row.MaxVariablesPerWorkflow,
		MaxAssertionsPerWorkflow: row.MaxAssertionsPerWorkflow,
		MaxWorkflowRunsPerMonth:    row.MaxWorkflowRunsPerMonth,
		MaxConcurrentRuns:          row.MaxConcurrentRuns,
		MinScheduleIntervalMinutes: row.MinScheduleIntervalMinutes,
		RunHistoryRetentionDays:    row.RunHistoryRetentionDays,
		MaxStepTimeoutSeconds:      row.MaxStepTimeoutSeconds,
		MaxRetryCountPerStep:       row.MaxRetryCountPerStep,
		MaxRequestBodySizeKB:       row.MaxRequestBodySizeKB,
		MaxResponseBodySizeKB:      row.MaxResponseBodySizeKB,
		AllowsOpenAPIImport:        row.AllowsOpenAPIImport,
		AllowsInsights:             row.AllowsInsights,
		AllowsDataExport:           row.AllowsDataExport,
		ExecutorPriority:           row.ExecutorPriority,
		CreatedAt:                  row.CreatedAt,
		UpdatedAt:                  row.UpdatedAt,
	}
}
