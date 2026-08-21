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
	ID                         uuid.UUID
	Name                       string
	MaxOrganizationMembers     int
	MaxWorkflows               int
	MaxStepsPerWorkflow        int
	MaxEndpoints               int
	MaxVariablesPerWorkflow    int
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
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
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
		MaxOrganizationMembers:     row.MaxOrganizationMembers,
		MaxWorkflows:               row.MaxWorkflows,
		MaxStepsPerWorkflow:        row.MaxStepsPerWorkflow,
		MaxEndpoints:               row.MaxEndpoints,
		MaxVariablesPerWorkflow:    row.MaxVariablesPerWorkflow,
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
