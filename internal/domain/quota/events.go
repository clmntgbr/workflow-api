package quota

import "time"

const (
	EventTypeQuotaThresholdReached = "quota.thresholdReached.v1"
	EventTypeQuotaExceeded         = "quota.exceeded.v1"

	QuotaNameWorkflowRuns   = "workflow_runs"
	QuotaNameConcurrentRuns = "concurrent_runs"
)

type QuotaThresholdReached struct {
	ID             string    `json:"eventId"`
	SubscriptionID string    `json:"subscriptionId"`
	UserID         string    `json:"userId"`
	QuotaName      string    `json:"quotaName"`
	Used           int64     `json:"used"`
	Max            int64     `json:"max"`
	Percent        int       `json:"percent"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e QuotaThresholdReached) EventID() string       { return e.ID }
func (e QuotaThresholdReached) EventType() string     { return EventTypeQuotaThresholdReached }
func (e QuotaThresholdReached) AggregateID() string   { return e.SubscriptionID }
func (e QuotaThresholdReached) OccurredAt() time.Time { return e.Timestamp }

type QuotaExceeded struct {
	ID             string    `json:"eventId"`
	SubscriptionID string    `json:"subscriptionId"`
	UserID         string    `json:"userId"`
	QuotaName      string    `json:"quotaName"`
	Used           int64     `json:"used"`
	Max            int64     `json:"max"`
	Timestamp      time.Time `json:"timestamp"`
}

func (e QuotaExceeded) EventID() string       { return e.ID }
func (e QuotaExceeded) EventType() string     { return EventTypeQuotaExceeded }
func (e QuotaExceeded) AggregateID() string   { return e.SubscriptionID }
func (e QuotaExceeded) OccurredAt() time.Time { return e.Timestamp }
