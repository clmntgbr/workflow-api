package mail

import (
	"fmt"
	"strings"
	"time"
)

func formatMoney(currency string, cents int64) string {
	if currency == "" {
		currency = "usd"
	}
	return fmt.Sprintf("%s %.2f", strings.ToUpper(currency), float64(cents)/100)
}

func formatTime(value *time.Time) string {
	if value == nil {
		return "—"
	}
	return value.UTC().Format("2006-01-02 15:04 UTC")
}

func formatDate(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format("2006-01-02")
}

func (s *mailURLs) billingURL() string {
	if s == nil || s.appBaseURL == "" {
		return ""
	}
	return s.appBaseURL + "/subscription"
}

type mailURLs struct {
	appBaseURL string
}

func newMailURLs(appBaseURL string) mailURLs {
	return mailURLs{appBaseURL: strings.TrimRight(appBaseURL, "/")}
}

func RecipientsFromUser(email string, banned bool) []string {
	if banned || email == "" {
		return nil
	}
	return []string{email}
}
