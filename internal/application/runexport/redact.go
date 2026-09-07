package runexport

import (
	"encoding/json"
	"strings"

	"go-api/internal/domain/httpquery"
)

func redactHeaders(headers map[string]string) map[string]string {
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

func redactBody(body any) any {
	switch typed := body.(type) {
	case map[string]any:
		return redactMap(typed)
	default:
		return body
	}
}

func redactQuery(query httpquery.Params) map[string]any {
	if len(query) == 0 {
		return map[string]any{}
	}
	out := make(map[string]any, len(query))
	for key, value := range query {
		items := value.Strings()
		if isSensitiveKey(key) {
			out[key] = "***"
			continue
		}
		if len(items) == 1 {
			out[key] = items[0]
			continue
		}
		out[key] = items
	}
	return out
}

func redactMap(in map[string]any) map[string]any {
	if len(in) == 0 {
		return map[string]any{}
	}
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

func toJSONCell(value any) string {
	if value == nil {
		return ""
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(encoded)
}
