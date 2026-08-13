package insight

import (
	"time"

	"github.com/google/uuid"
)

type Insight struct {
	ID        uuid.UUID
	StepRunID uuid.UUID

	StartTime *time.Time
	EndTime   *time.Time
	QueueTime *time.Duration

	DNSLookupDuration *time.Duration
	TCPConnectionTime *time.Duration
	TLSHandshakeTime  *time.Duration
	TTFB              *time.Duration

	Duration      *time.Duration
	StatusCode    *int
	ResponseSize  *int64
	RequestSize   *int64
	AttemptNumber int
	TotalAttempts int
	ErrorMessage  string
	ErrorType     string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type NewInsightParams struct {
	StepRunID         uuid.UUID
	StartTime         time.Time
	EndTime           time.Time
	QueueTime         time.Duration
	DNSLookupDuration time.Duration
	TCPConnectionTime time.Duration
	TLSHandshakeTime  time.Duration
	TTFB              time.Duration
	Duration          time.Duration
	StatusCode        int
	ResponseSize      int64
	RequestSize       int64
	AttemptNumber     int
	TotalAttempts     int
	ErrorMessage      string
	ErrorType         string
	HasStatusCode     bool
}

func NewInsight(p NewInsightParams) *Insight {
	now := time.Now().UTC()
	start := p.StartTime.UTC()
	end := p.EndTime.UTC()

	insight := &Insight{
		ID:            uuid.New(),
		StepRunID:     p.StepRunID,
		StartTime:     &start,
		EndTime:       &end,
		AttemptNumber: p.AttemptNumber,
		TotalAttempts: p.TotalAttempts,
		ErrorMessage:  p.ErrorMessage,
		ErrorType:     p.ErrorType,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	insight.QueueTime = durationPtr(p.QueueTime)
	insight.DNSLookupDuration = durationPtr(p.DNSLookupDuration)
	insight.TCPConnectionTime = durationPtr(p.TCPConnectionTime)
	insight.TLSHandshakeTime = durationPtr(p.TLSHandshakeTime)
	insight.TTFB = durationPtr(p.TTFB)
	insight.Duration = durationPtr(p.Duration)
	insight.ResponseSize = int64Ptr(p.ResponseSize)
	insight.RequestSize = int64Ptr(p.RequestSize)
	if p.HasStatusCode {
		code := p.StatusCode
		insight.StatusCode = &code
	}

	return insight
}

func durationPtr(d time.Duration) *time.Duration {
	if d < 0 {
		d = 0
	}
	return &d
}

func int64Ptr(v int64) *int64 {
	return &v
}
