package httpexecutor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-api/internal/domain/port"
)

type Client struct {
	httpClient *http.Client
}

func New() *Client {
	return &Client{
		httpClient: &http.Client{},
	}
}

func (c *Client) Do(ctx context.Context, req port.HTTPRequest) (*port.HTTPResponse, error) {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	reqCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	target, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}
	query := target.Query()
	for key, value := range req.Query {
		query.Set(key, value)
	}
	target.RawQuery = query.Encode()

	var bodyReader io.Reader
	if req.Body != nil && len(req.Body) > 0 {
		encoded, err := json.Marshal(req.Body)
		if err != nil {
			return nil, fmt.Errorf("encode body: %w", err)
		}
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

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	rawBody, err := io.ReadAll(io.LimitReader(httpResp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

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
	}, nil
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
