package read

import (
	"context"
	"time"

	domaininsight "go-api/internal/domain/insight"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type insightRow struct {
	ID                uuid.UUID
	StepRunID         uuid.UUID
	StartTime         *time.Time
	EndTime           *time.Time
	QueueTimeNs       *int64 `gorm:"column:queue_time_ns"`
	DNSLookupDuration *int64 `gorm:"column:dns_lookup_duration_ns"`
	TCPConnectionTime *int64 `gorm:"column:tcp_connection_time_ns"`
	TLSHandshakeTime  *int64 `gorm:"column:tls_handshake_time_ns"`
	TTFBNs            *int64 `gorm:"column:ttfb_ns"`
	DurationNs        *int64 `gorm:"column:duration_ns"`
	StatusCode        *int
	ResponseSize      *int64
	RequestSize       *int64
	AttemptNumber     int
	TotalAttempts     int
	ErrorMessage      string
	ErrorType         string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func (insightRow) TableName() string { return "insights" }

type insightReadRepository struct {
	db *gorm.DB
}

func NewInsightReadRepository(db *gorm.DB) domaininsight.InsightReadRepository {
	return &insightReadRepository{db: db}
}

func (r *insightReadRepository) FindByStepRunID(
	ctx context.Context,
	stepRunID uuid.UUID,
) ([]domaininsight.InsightView, error) {
	return r.find(ctx, []uuid.UUID{stepRunID})
}

func (r *insightReadRepository) FindByStepRunIDs(
	ctx context.Context,
	stepRunIDs []uuid.UUID,
) ([]domaininsight.InsightView, error) {
	return r.find(ctx, stepRunIDs)
}

func (r *insightReadRepository) find(
	ctx context.Context,
	stepRunIDs []uuid.UUID,
) ([]domaininsight.InsightView, error) {
	if len(stepRunIDs) == 0 {
		return []domaininsight.InsightView{}, nil
	}

	var rows []insightRow
	err := r.db.WithContext(ctx).
		Where("step_run_id IN ?", stepRunIDs).
		Order("attempt_number ASC, created_at ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	views := make([]domaininsight.InsightView, 0, len(rows))
	for _, row := range rows {
		views = append(views, toInsightView(row))
	}
	return views, nil
}

func toInsightView(row insightRow) domaininsight.InsightView {
	return domaininsight.InsightView{
		ID:                row.ID,
		StepRunID:         row.StepRunID,
		StartTime:         row.StartTime,
		EndTime:           row.EndTime,
		QueueTime:         insightNsToDuration(row.QueueTimeNs),
		DNSLookupDuration: insightNsToDuration(row.DNSLookupDuration),
		TCPConnectionTime: insightNsToDuration(row.TCPConnectionTime),
		TLSHandshakeTime:  insightNsToDuration(row.TLSHandshakeTime),
		TTFB:              insightNsToDuration(row.TTFBNs),
		Duration:          insightNsToDuration(row.DurationNs),
		StatusCode:        row.StatusCode,
		ResponseSize:      row.ResponseSize,
		RequestSize:       row.RequestSize,
		AttemptNumber:     row.AttemptNumber,
		TotalAttempts:     row.TotalAttempts,
		ErrorMessage:      row.ErrorMessage,
		ErrorType:         row.ErrorType,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
	}
}

func insightNsToDuration(ns *int64) *time.Duration {
	if ns == nil {
		return nil
	}
	d := time.Duration(*ns)
	return &d
}
