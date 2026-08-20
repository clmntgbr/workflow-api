package httpexecutor

import (
	"encoding/json"
	"log"
	"strings"

	"go-api/internal/domain/httpquery"
)

func logHTTPRequest(
	method string,
	fullURL string,
	query httpquery.Params,
	headers map[string]string,
	body map[string]any,
) {
	payload := map[string]any{
		"method":  method,
		"url":     fullURL,
		"query":   queryForLog(query),
		"headers": headersForLog(headers),
		"body":    bodyForLog(body),
	}
	encoded, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		log.Printf("httpexecutor request method=%s url=%s (failed to encode log payload: %v)", method, fullURL, err)
		return
	}
	log.Printf("httpexecutor request\n%s", string(encoded))
}

func queryForLog(query httpquery.Params) any {
	if len(query) == 0 {
		return map[string]any{}
	}
	return query
}

func headersForLog(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return map[string]string{}
	}
	out := make(map[string]string, len(headers))
	for key, value := range headers {
		if isSensitiveHeader(key) {
			out[key] = redactSecret(value)
			continue
		}
		out[key] = value
	}
	return out
}

func bodyForLog(body map[string]any) map[string]any {
	if len(body) == 0 {
		return map[string]any{}
	}
	return redactMap(body)
}

func redactMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for key, value := range in {
		if isSensitiveKey(key) {
			if asString, ok := value.(string); ok {
				out[key] = redactSecret(asString)
			} else {
				out[key] = "***"
			}
			continue
		}
		switch typed := value.(type) {
		case map[string]any:
			out[key] = redactMap(typed)
		default:
			out[key] = value
		}
	}
	return out
}

func isSensitiveHeader(key string) bool {
	switch strings.ToLower(strings.TrimSpace(key)) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key", "api-key":
		return true
	default:
		return false
	}
}

func isSensitiveKey(key string) bool {
	normalized := strings.ToLower(strings.TrimSpace(key))
	switch normalized {
	case "password", "passwd", "secret", "token", "access_token", "refresh_token", "client_secret", "authorization":
		return true
	default:
		return strings.Contains(normalized, "password") || strings.Contains(normalized, "secret") || strings.Contains(normalized, "token")
	}
}

func redactSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return "Bearer ***"
	}
	if len(value) <= 8 {
		return "***"
	}
	return value[:4] + "***" + value[len(value)-2:]
}
