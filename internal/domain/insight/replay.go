package insight

import (
	"time"

	"github.com/google/uuid"
)

func (v InsightView) ReplayOnto(stepRunID uuid.UUID) *Insight {
	now := time.Now().UTC()
	return &Insight{
		ID:                uuid.New(),
		StepRunID:         stepRunID,
		StartTime:         cloneTime(v.StartTime),
		EndTime:           cloneTime(v.EndTime),
		QueueTime:         cloneDuration(v.QueueTime),
		DNSLookupDuration: cloneDuration(v.DNSLookupDuration),
		TCPConnectionTime: cloneDuration(v.TCPConnectionTime),
		TLSHandshakeTime:  cloneDuration(v.TLSHandshakeTime),
		TTFB:              cloneDuration(v.TTFB),
		Duration:          cloneDuration(v.Duration),
		StatusCode:        cloneInt(v.StatusCode),
		ResponseSize:      cloneInt64(v.ResponseSize),
		RequestSize:       cloneInt64(v.RequestSize),
		AttemptNumber:     v.AttemptNumber,
		TotalAttempts:     v.TotalAttempts,
		ErrorMessage:      v.ErrorMessage,
		ErrorType:         v.ErrorType,
		CreatedAt:         now,
		UpdatedAt:         now,
	}
}

func cloneTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := value.UTC()
	return &copied
}

func cloneDuration(value *time.Duration) *time.Duration {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}
