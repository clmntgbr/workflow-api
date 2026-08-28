package activitylog

import (
	"fmt"
	"strings"
)

func httpCallLabel(method, url string) string {
	method = strings.ToUpper(strings.TrimSpace(method))
	url = strings.TrimSpace(url)
	if method == "" && url == "" {
		return ""
	}
	if method == "" {
		return url
	}
	if url == "" {
		return method
	}
	return fmt.Sprintf("%s %s", method, url)
}

func onAttemptPhrase(attempt, maxAttempts int) string {
	if attempt <= 0 {
		return ""
	}
	if maxAttempts > 0 {
		return fmt.Sprintf("on attempt %d of %d", attempt, maxAttempts)
	}
	return fmt.Sprintf("on attempt %d", attempt)
}

func withHTTPCallPhrase(method, url string) string {
	call := httpCallLabel(method, url)
	if call == "" {
		return ""
	}
	return fmt.Sprintf("with %s", call)
}

func withHTTPStatusPhrase(statusCode int) string {
	if statusCode <= 0 {
		return ""
	}
	return fmt.Sprintf("with HTTP %d", statusCode)
}

func triggeredByLabel(triggeredBy string) string {
	switch strings.TrimSpace(triggeredBy) {
	case "user":
		return "manual"
	case "schedule":
		return "schedule"
	case "api":
		return "API"
	case "cli":
		return "CLI"
	case "webhook":
		return "webhook"
	default:
		if triggeredBy == "" {
			return "unknown"
		}
		return triggeredBy
	}
}

func scheduledSkipReasonLabel(reason string) string {
	switch reason {
	case "already_in_progress":
		return "a run is already in progress"
	case "quota_exceeded":
		return "quota exceeded"
	default:
		return reason
	}
}
