package port

import (
	"context"
	"time"
)

type HTTPRequest struct {
	Method  string
	URL     string
	Headers map[string]string
	Query   map[string]string
	Body    map[string]any
	Timeout time.Duration
}

type HTTPResponse struct {
	Status  int
	Headers map[string]string
	Body    any
}

type HTTPExecutor interface {
	Do(ctx context.Context, req HTTPRequest) (*HTTPResponse, error)
}
