package mail

import (
	"context"
	"fmt"

	domaininvoice "go-api/internal/domain/invoice"
	"go-api/internal/domain/port"
	domainsubscription "go-api/internal/domain/subscription"
)

type BillingMailService struct {
	sender port.MailSender
	urls   mailURLs
}

func NewBillingMailService(sender port.MailSender, appBaseURL string) *BillingMailService {
	return &BillingMailService{
		sender: sender,
		urls:   newMailURLs(appBaseURL),
	}
}

func (s *BillingMailService) NotifyPaymentSucceeded(ctx context.Context, evt domaininvoice.InvoicePaymentSucceeded, to []string) error {
	if len(to) == 0 {
		return nil
	}
	amount := formatMoney(evt.Currency, evt.AmountPaid)
	invoiceURL := evt.HostedInvoiceURL
	if invoiceURL == "" {
		invoiceURL = s.urls.billingURL()
	}
	return s.sender.Send(ctx, port.SendMailInput{
		To:           to,
		Subject:      fmt.Sprintf("Payment received%s", subjectAmount(amount)),
		TemplateName: "invoice_payment_succeeded",
		TemplateData: map[string]any{
			"Preheader":     fmt.Sprintf("We received your payment%s.", subjectAmount(amount)),
			"Amount":        amount,
			"InvoiceNumber": evt.Number,
			"PaidAt":        formatTime(evt.PaidAt),
			"InvoiceURL":    invoiceURL,
		},
	})
}

func (s *BillingMailService) NotifyPaymentFailed(ctx context.Context, evt domaininvoice.InvoicePaymentFailed, to []string) error {
	if len(to) == 0 {
		return nil
	}
	amount := formatMoney(evt.Currency, evt.AmountDue)
	return s.sender.Send(ctx, port.SendMailInput{
		To:           to,
		Subject:      "Payment failed — update your billing details",
		TemplateName: "invoice_payment_failed",
		TemplateData: map[string]any{
			"Preheader":     "Update your payment method to keep your subscription active.",
			"Amount":        amount,
			"InvoiceNumber": evt.Number,
			"AttemptCount":  evt.AttemptCount,
			"BillingURL":    firstNonEmpty(evt.HostedInvoiceURL, s.urls.billingURL()),
		},
	})
}

func (s *BillingMailService) NotifyPlanChanged(ctx context.Context, previousName, newName string, to []string) error {
	if len(to) == 0 || previousName == "" || newName == "" {
		return nil
	}
	kind := "changed"
	if previousName != newName {
		kind = "updated"
	}
	return s.sender.Send(ctx, port.SendMailInput{
		To:           to,
		Subject:      fmt.Sprintf("Your plan is now %s", newName),
		TemplateName: "subscription_plan_changed",
		TemplateData: map[string]any{
			"Preheader":        fmt.Sprintf("Your plan is now %s.", newName),
			"ChangeKind":       kind,
			"PreviousPlanName": previousName,
			"PlanName":         newName,
			"BillingURL":       s.urls.billingURL(),
		},
	})
}

func (s *BillingMailService) NotifyRenewalUpcoming(ctx context.Context, evt domainsubscription.SubscriptionRenewalUpcoming, planName string, to []string) error {
	if len(to) == 0 {
		return nil
	}
	if planName == "" {
		planName = "Workflow"
	}
	amount := formatMoney(evt.Currency, evt.AmountDue)
	return s.sender.Send(ctx, port.SendMailInput{
		To:           to,
		Subject:      fmt.Sprintf("Your %s subscription renews soon", planName),
		TemplateName: "subscription_renewal_upcoming",
		TemplateData: map[string]any{
			"Preheader":  fmt.Sprintf("Your %s subscription renews soon.", planName),
			"PlanName":   planName,
			"Amount":     amount,
			"PeriodEnd":  formatDate(evt.PeriodEnd),
			"BillingURL": s.urls.billingURL(),
		},
	})
}

func (s *BillingMailService) NotifyPaymentMethodExpiring(ctx context.Context, evt domainsubscription.SubscriptionPaymentMethodExpiring, to []string) error {
	if len(to) == 0 {
		return nil
	}
	expiresOn := ""
	if evt.ExpMonth > 0 && evt.ExpYear > 0 {
		expiresOn = fmt.Sprintf("%02d/%d", evt.ExpMonth, evt.ExpYear)
	}
	return s.sender.Send(ctx, port.SendMailInput{
		To:           to,
		Subject:      "Your payment card is about to expire",
		TemplateName: "payment_method_expiring",
		TemplateData: map[string]any{
			"Preheader":  "Update your card to keep your subscription active.",
			"Brand":      evt.Brand,
			"Last4":      evt.Last4,
			"ExpiresOn":  expiresOn,
			"BillingURL": s.urls.billingURL(),
		},
	})
}

func subjectAmount(amount string) string {
	if amount == "" {
		return ""
	}
	return " — " + amount
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
