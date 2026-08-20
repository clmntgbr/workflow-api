package port

import (
	"context"
	"time"

	"go-api/internal/domain/httpquery"
)

type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Query   httpquery.Params
	Body    map[string]any
	Timeout time.Duration
}

type HTTPTiming struct {
	StartTime         time.Time
	EndTime           time.Time
	DNSLookupDuration time.Duration
	TCPConnectionTime time.Duration
	TLSHandshakeTime  time.Duration
	TTFB              time.Duration
	Duration          time.Duration
	RequestSize       int64
	ResponseSize      int64
	ErrorType         string
}

type HTTPResponse struct {
	Status  int
	Headers map[string]string
	Body    any
	Timing  HTTPTiming
}

type HTTPExecutor interface {
	Do(ctx context.Context, req HTTPRequest) (*HTTPResponse, error)
}
