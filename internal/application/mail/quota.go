package mail

import (
	"context"
	"fmt"

	"go-api/internal/domain/port"
	domainquota "go-api/internal/domain/quota"
)

type QuotaMailService struct {
	sender port.MailSender
	urls   mailURLs
}

func NewQuotaMailService(sender port.MailSender, appBaseURL string) *QuotaMailService {
	return &QuotaMailService{
		sender: sender,
		urls:   newMailURLs(appBaseURL),
	}
}

func (s *QuotaMailService) NotifyThresholdReached(ctx context.Context, evt domainquota.QuotaThresholdReached, to []string) error {
	if len(to) == 0 {
		return nil
	}
	return s.sender.Send(ctx, port.SendMailInput{
		To:           to,
		Subject:      fmt.Sprintf("You've used %d%% of your workflow run quota", evt.Percent),
		TemplateName: "quota_threshold_reached",
		TemplateData: map[string]any{
			"Preheader":  fmt.Sprintf("You have used %d%% of your workflow run quota.", evt.Percent),
			"Used":       evt.Used,
			"Max":        evt.Max,
			"Percent":    evt.Percent,
			"BillingURL": s.urls.billingURL(),
		},
	})
}

func (s *QuotaMailService) NotifyExceeded(ctx context.Context, evt domainquota.QuotaExceeded, to []string) error {
	if len(to) == 0 {
		return nil
	}
	return s.sender.Send(ctx, port.SendMailInput{
		To:           to,
		Subject:      "An action was blocked by your plan quota",
		TemplateName: "quota_exceeded",
		TemplateData: map[string]any{
			"Preheader":  "An action was blocked because you reached a plan limit.",
			"QuotaLabel": quotaLabel(evt.QuotaName),
			"Used":       evt.Used,
			"Max":        evt.Max,
			"BillingURL": s.urls.billingURL(),
		},
	})
}

func quotaLabel(name string) string {
	switch name {
	case domainquota.QuotaNameConcurrentRuns:
		return "concurrent runs"
	default:
		return "monthly workflow runs"
	}
}
