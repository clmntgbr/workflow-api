package httpexecutor

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"time"

	"go-api/internal/domain/httpquery"
	"go-api/internal/domain/port"
)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{
			// Timeout is applied per request via context.
			Timeout: 0,
		},
	}
}

func (c *Client) Do(ctx context.Context, req port.HTTPRequest) (*port.HTTPResponse, error) {
	reqCtx, cancel := context.WithTimeout(ctx, req.Timeout)
	defer cancel()

	target, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	target.RawQuery = httpquery.BuildQueryString(req.Query)

	var bodyBytes []byte
	var bodyReader io.Reader
	if req.Body != nil && len(req.Body) > 0 {
		encoded, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("encode body: %w", err)
		}
		bodyBytes = encoded
		bodyReader = bytes.NewReader(encoded)
	}

	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = http.MethodGet
	}

	httpReq, err := http.NewRequestWithContext(reqCtx, method, target.String(), bodyReader)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}
	if bodyReader != nil && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
	logHTTPRequest(method, target.String(), req.Query, req.Headers, req.Body)

	timing := &requestTiming{start: time.Now().UTC()}
	trace := &httptrace.ClientTrace{
		DNSStart: func(httptrace.DNSStartInfo) {
			timing.dnsStart = time.Now().UTC()
		},
		DNSDone: func(httptrace.DNSDoneInfo) {
			if !timing.dnsStart.IsZero() {
				timing.dns = time.Since(timing.dnsStart)
			}
		},
		ConnectStart: func(_, _ string) {
			timing.tcpStart = time.Now().UTC()
		},
		ConnectDone: func(_, _ string, err error) {
			if !timing.tcpStart.IsZero() {
				timing.tcp = time.Since(timing.tcpStart)
			}
			if err != nil {
				timing.connectErr = err
			}
		},
		TLSHandshakeStart: func() {
			timing.tlsStart = time.Now().UTC()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if !timing.tlsStart.IsZero() {
				timing.tls = time.Since(timing.tlsStart)
			}
			if err != nil {
				timing.tlsErr = err
			}
		},
		GotFirstResponseByte: func() {
			if !timing.start.IsZero() {
				timing.ttfb = time.Since(timing.start)
			}
		},
	}
	httpReq = httpReq.WithContext(httptrace.WithClientTrace(httpReq.Context(), trace))

	httpResp, err := c.httpClient.Do(httpReq)
	end := time.Now().UTC()
	baseTiming := port.HTTPTiming{
		StartTime:         timing.start,
		EndTime:           end,
		DNSLookupDuration: timing.dns,
		TCPConnectionTime: timing.tcp,
		TLSHandshakeTime:  timing.tls,
		TTFB:              timing.ttfb,
		Duration:          end.Sub(timing.start),
		RequestSize:       int64(len(bodyBytes)),
	}

	if err != nil {
		baseTiming.ErrorType = classifyError(err, timing)
		return &port.HTTPResponse{Timing: baseTiming}, err
	}
	defer httpResp.Body.Close()

	rawBody, err := io.ReadAll(io.LimitReader(httpResp.Body, 8<<20))
	if err != nil {
		baseTiming.EndTime = time.Now().UTC()
		baseTiming.Duration = baseTiming.EndTime.Sub(timing.start)
		baseTiming.ErrorType = "response_read"
		return &port.HTTPResponse{Timing: baseTiming}, fmt.Errorf("read response: %w", err)
	}
	baseTiming.ResponseSize = int64(len(rawBody))
	baseTiming.EndTime = time.Now().UTC()
	baseTiming.Duration = baseTiming.EndTime.Sub(timing.start)

	headers := make(map[string]string, len(httpResp.Header))
	for key, values := range httpResp.Header {
		if len(values) == 0 {
			continue
		}
		headers[key] = values[0]
	}

	return &port.HTTPResponse{
		Status:  httpResp.StatusCode,
		Headers: headers,
		Body:    decodeBody(rawBody),
		Timing:  baseTiming,
	}, nil
}

type requestTiming struct {
	start      time.Time
	dnsStart   time.Time
	tcpStart   time.Time
	tlsStart   time.Time
	dns        time.Duration
	tcp        time.Duration
	tls        time.Duration
	ttfb       time.Duration
	connectErr error
	tlsErr     error
}

func classifyError(err error, timing *requestTiming) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return "timeout"
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns"
	}
	if timing != nil && timing.tlsErr != nil {
		return "tls"
	}
	if timing != nil && timing.connectErr != nil {
		return "connection"
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return "timeout"
		}
		return "network"
	}
	return "unknown"
}

func decodeBody(raw []byte) any {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil
	}
	var asJSON any
	if err := json.Unmarshal(trimmed, &asJSON); err == nil {
		return asJSON
	}
	return string(raw)
}
