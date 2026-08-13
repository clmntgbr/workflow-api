package insight

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type InsightWriteRepository interface {
	Save(ctx context.Context, insight *Insight) error
}

type InsightReadRepository interface {
	FindByStepRunID(ctx context.Context, stepRunID uuid.UUID) ([]InsightView, error)
	FindByStepRunIDs(ctx context.Context, stepRunIDs []uuid.UUID) ([]InsightView, error)
}

type InsightView struct {
	ID                uuid.UUID
	StepRunID         uuid.UUID
	StartTime         *time.Time
	EndTime           *time.Time
	QueueTime         *time.Duration
	DNSLookupDuration *time.Duration
	TCPConnectionTime *time.Duration
	TLSHandshakeTime  *time.Duration
	TTFB              *time.Duration
	Duration          *time.Duration
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
