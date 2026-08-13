package write

import (
	"time"

	domaininsight "go-api/internal/domain/insight"

	"github.com/google/uuid"
)

type InsightModel struct {
	ID                uuid.UUID  `gorm:"column:id;primaryKey"`
	StepRunID         uuid.UUID  `gorm:"column:step_run_id"`
	StartTime         *time.Time `gorm:"column:start_time"`
	EndTime           *time.Time `gorm:"column:end_time"`
	QueueTimeNs       *int64     `gorm:"column:queue_time_ns"`
	DNSLookupDuration *int64     `gorm:"column:dns_lookup_duration_ns"`
	TCPConnectionTime *int64     `gorm:"column:tcp_connection_time_ns"`
	TLSHandshakeTime  *int64     `gorm:"column:tls_handshake_time_ns"`
	TTFBNs            *int64     `gorm:"column:ttfb_ns"`
	DurationNs        *int64     `gorm:"column:duration_ns"`
	StatusCode        *int       `gorm:"column:status_code"`
	ResponseSize      *int64     `gorm:"column:response_size"`
	RequestSize       *int64     `gorm:"column:request_size"`
	AttemptNumber     int        `gorm:"column:attempt_number"`
	TotalAttempts     int        `gorm:"column:total_attempts"`
	ErrorMessage      string     `gorm:"column:error_message"`
	ErrorType         string     `gorm:"column:error_type"`
	CreatedAt         time.Time  `gorm:"column:created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at"`
}

func (InsightModel) TableName() string {
	return "insights"
}

func insightModelFromDomain(i *domaininsight.Insight) *InsightModel {
	return &InsightModel{
		ID:                i.ID,
		StepRunID:         i.StepRunID,
		StartTime:         i.StartTime,
		EndTime:           i.EndTime,
		QueueTimeNs:       durationToNs(i.QueueTime),
		DNSLookupDuration: durationToNs(i.DNSLookupDuration),
		TCPConnectionTime: durationToNs(i.TCPConnectionTime),
		TLSHandshakeTime:  durationToNs(i.TLSHandshakeTime),
		TTFBNs:            durationToNs(i.TTFB),
		DurationNs:        durationToNs(i.Duration),
		StatusCode:        i.StatusCode,
		ResponseSize:      i.ResponseSize,
		RequestSize:       i.RequestSize,
		AttemptNumber:     i.AttemptNumber,
		TotalAttempts:     i.TotalAttempts,
		ErrorMessage:      i.ErrorMessage,
		ErrorType:         i.ErrorType,
		CreatedAt:         i.CreatedAt,
		UpdatedAt:         i.UpdatedAt,
	}
}

func durationToNs(d *time.Duration) *int64 {
	if d == nil {
		return nil
	}
	ns := d.Nanoseconds()
	return &ns
}
