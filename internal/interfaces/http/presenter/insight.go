package presenter

import (
	"time"

	domaininsight "go-api/internal/domain/insight"
)

type InsightResponse struct {
	ID                string     `json:"id"`
	StartTime         *time.Time `json:"startTime"`
	EndTime           *time.Time `json:"endTime"`
	QueueTime         *int64     `json:"queueTime"`
	DNSLookupDuration *int64     `json:"dnsLookupDuration"`
	TCPConnectionTime *int64     `json:"tcpConnectionTime"`
	TLSHandshakeTime  *int64     `json:"tlsHandshakeTime"`
	TTFB              *int64     `json:"ttfb"`
	Duration          *int64     `json:"duration"`
	StatusCode        *int       `json:"statusCode"`
	ResponseSize      *int64     `json:"responseSize"`
	RequestSize       *int64     `json:"requestSize"`
	AttemptNumber     int        `json:"attemptNumber"`
	TotalAttempts     int        `json:"totalAttempts"`
	ErrorMessage      *string    `json:"errorMessage"`
	ErrorType         *string    `json:"errorType"`
}

func NewInsightResponseFromView(view domaininsight.InsightView) InsightResponse {
	return InsightResponse{
		ID:                view.ID.String(),
		StartTime:         view.StartTime,
		EndTime:           view.EndTime,
		QueueTime:         durationToMs(view.QueueTime),
		DNSLookupDuration: durationToMs(view.DNSLookupDuration),
		TCPConnectionTime: durationToMs(view.TCPConnectionTime),
		TLSHandshakeTime:  durationToMs(view.TLSHandshakeTime),
		TTFB:              durationToMs(view.TTFB),
		Duration:          durationToMs(view.Duration),
		StatusCode:        view.StatusCode,
		ResponseSize:      view.ResponseSize,
		RequestSize:       view.RequestSize,
		AttemptNumber:     view.AttemptNumber,
		TotalAttempts:     view.TotalAttempts,
		ErrorMessage:      optionalNonEmptyString(view.ErrorMessage),
		ErrorType:         optionalNonEmptyString(view.ErrorType),
	}
}

func NewInsightListResponseFromViews(views []domaininsight.InsightView) []InsightResponse {
	if len(views) == 0 {
		return nil
	}
	items := make([]InsightResponse, 0, len(views))
	for _, view := range views {
		items = append(items, NewInsightResponseFromView(view))
	}
	return items
}

func durationToMs(d *time.Duration) *int64 {
	if d == nil {
		return nil
	}
	ms := d.Milliseconds()
	return &ms
}
