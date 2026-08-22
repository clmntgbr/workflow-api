package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"sync"
	"time"

	cmdsubscription "go-api/internal/application/command/subscription"
	infrastripe "go-api/internal/infrastructure/stripe"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
)

type BillingWebhookHandler struct {
	mu                             sync.Mutex
	checkoutCompletedHandler       *cmdsubscription.CheckoutCompletedHandler
	subscriptionUpdatedHandler     *cmdsubscription.SubscriptionUpdatedHandler
	subscriptionDeletedHandler     *cmdsubscription.SubscriptionDeletedHandler
	invoicePaymentSucceededHandler *cmdsubscription.InvoicePaymentSucceededHandler
	invoicePaymentFailedHandler    *cmdsubscription.InvoicePaymentFailedHandler
	upsertInvoiceHandler           *cmdsubscription.UpsertInvoiceHandler
}

func NewBillingWebhookHandler(
	checkoutCompletedHandler *cmdsubscription.CheckoutCompletedHandler,
	subscriptionUpdatedHandler *cmdsubscription.SubscriptionUpdatedHandler,
	subscriptionDeletedHandler *cmdsubscription.SubscriptionDeletedHandler,
	invoicePaymentSucceededHandler *cmdsubscription.InvoicePaymentSucceededHandler,
	invoicePaymentFailedHandler *cmdsubscription.InvoicePaymentFailedHandler,
	upsertInvoiceHandler *cmdsubscription.UpsertInvoiceHandler,
) *BillingWebhookHandler {
	return &BillingWebhookHandler{
		checkoutCompletedHandler:       checkoutCompletedHandler,
		subscriptionUpdatedHandler:     subscriptionUpdatedHandler,
		subscriptionDeletedHandler:     subscriptionDeletedHandler,
		invoicePaymentSucceededHandler: invoicePaymentSucceededHandler,
		invoicePaymentFailedHandler:    invoicePaymentFailedHandler,
		upsertInvoiceHandler:           upsertInvoiceHandler,
	}
}

func (h *BillingWebhookHandler) Execute(c fiber.Ctx) error {
	event := c.Locals("payload").(stripe.Event)
	log.Printf("stripe webhook: received event id=%s type=%s", event.ID, event.Type)

	h.mu.Lock()
	defer h.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Printf("stripe webhook: processing event id=%s type=%s", event.ID, event.Type)
	if err := h.dispatch(ctx, event); err != nil {
		log.Printf("stripe webhook: failed event id=%s type=%s: %v", event.ID, event.Type, err)
		if errors.Is(err, cmdsubscription.ErrStripeSubscriptionNotLinked) {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "subscription not linked yet",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to process event",
		})
	}

	log.Printf("stripe webhook: processed event id=%s type=%s", event.ID, event.Type)
	return c.SendStatus(fiber.StatusOK)
}

func (h *BillingWebhookHandler) dispatch(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "checkout.session.completed":
		return h.handleCheckoutCompleted(ctx, event)
	case "customer.subscription.updated":
		return h.handleSubscriptionUpdated(ctx, event)
	case "customer.subscription.deleted":
		return h.handleSubscriptionDeleted(ctx, event)
	case "invoice.payment_succeeded":
		return h.handleInvoicePaymentSucceeded(ctx, event)
	case "invoice.payment_failed":
		return h.handleInvoicePaymentFailed(ctx, event)
	default:
		log.Printf("stripe webhook: ignoring unhandled event id=%s type=%s", event.ID, event.Type)
		return nil
	}
}

func (h *BillingWebhookHandler) handleCheckoutCompleted(ctx context.Context, event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return err
	}

	userID, err := uuid.Parse(session.ClientReferenceID)
	if err != nil {
		return err
	}

	cmd := cmdsubscription.CheckoutCompletedCommand{
		UserID: userID,
	}
	if session.Customer != nil {
		cmd.StripeCustomerID = session.Customer.ID
	}
	if session.Subscription != nil {
		cmd.StripeSubscriptionID = session.Subscription.ID
	}

	return h.checkoutCompletedHandler.Handle(ctx, cmd)
}

func (h *BillingWebhookHandler) handleSubscriptionUpdated(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	data := infrastripe.ExtractSubscriptionData(&sub)

	return h.subscriptionUpdatedHandler.Handle(ctx, cmdsubscription.SubscriptionUpdatedCommand{
		StripeSubscriptionID: data.ID,
		StripeCustomerID:     data.CustomerID,
		StripePriceID:        data.PriceID,
		Status:               data.Status,
		CancelAtPeriodEnd:    data.CancelAtPeriodEnd,
		CurrentPeriodStart:   data.CurrentPeriodStart,
		CurrentPeriodEnd:     data.CurrentPeriodEnd,
	})
}

func (h *BillingWebhookHandler) handleSubscriptionDeleted(ctx context.Context, event stripe.Event) error {
	var sub stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &sub); err != nil {
		return err
	}

	return h.subscriptionDeletedHandler.Handle(ctx, cmdsubscription.SubscriptionDeletedCommand{
		StripeSubscriptionID: sub.ID,
	})
}

func (h *BillingWebhookHandler) handleInvoicePaymentSucceeded(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		log.Printf("stripe webhook: failed to unmarshal invoice.payment_succeeded: %v", err)
		return err
	}

	upsertCmd := upsertInvoiceCommandFromStripe(&invoice)
	log.Printf(
		"stripe webhook: invoice.payment_succeeded mapped invoiceID=%s subscriptionID=%s customerID=%s parent=%v",
		upsertCmd.StripeInvoiceID,
		upsertCmd.StripeSubscriptionID,
		upsertCmd.StripeCustomerID,
		invoice.Parent != nil,
	)

	if err := h.upsertInvoiceHandler.Handle(ctx, upsertCmd); err != nil {
		return err
	}

	return h.invoicePaymentSucceededHandler.Handle(ctx, cmdsubscription.InvoicePaymentSucceededCommand{
		StripeSubscriptionID: upsertCmd.StripeSubscriptionID,
		StripeCustomerID:     upsertCmd.StripeCustomerID,
		BillingReason:        string(invoice.BillingReason),
		PeriodStart:          unixToTime(invoice.PeriodStart),
		PeriodEnd:            unixToTime(invoice.PeriodEnd),
	})
}

func (h *BillingWebhookHandler) handleInvoicePaymentFailed(ctx context.Context, event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return err
	}

	upsertCmd := upsertInvoiceCommandFromStripe(&invoice)
	if err := h.upsertInvoiceHandler.Handle(ctx, upsertCmd); err != nil {
		return err
	}

	return h.invoicePaymentFailedHandler.Handle(ctx, cmdsubscription.InvoicePaymentFailedCommand{
		StripeSubscriptionID: upsertCmd.StripeSubscriptionID,
		StripeCustomerID:     upsertCmd.StripeCustomerID,
	})
}

func upsertInvoiceCommandFromStripe(invoice *stripe.Invoice) cmdsubscription.UpsertInvoiceCommand {
	cmd := cmdsubscription.UpsertInvoiceCommand{
		StripeInvoiceID:      invoice.ID,
		StripeCustomerID:     customerIDFromInvoice(invoice),
		StripeSubscriptionID: subscriptionIDFromInvoice(invoice),
		Number:               invoice.Number,
		Status:               string(invoice.Status),
		Currency:             string(invoice.Currency),
		AmountDue:            invoice.AmountDue,
		AmountPaid:           invoice.AmountPaid,
		Total:                invoice.Total,
		HostedInvoiceURL:     invoice.HostedInvoiceURL,
		InvoicePDF:           invoice.InvoicePDF,
		BillingReason:        string(invoice.BillingReason),
		Description:          invoiceDescription(invoice),
		AttemptCount:         invoice.AttemptCount,
		PeriodStart:          unixToTime(invoice.PeriodStart),
		PeriodEnd:            unixToTime(invoice.PeriodEnd),
		StripeCreatedAt:      unixToTime(invoice.Created),
	}

	if invoice.StatusTransitions != nil {
		if paidAt := unixToTime(invoice.StatusTransitions.PaidAt); !paidAt.IsZero() {
			cmd.PaidAt = &paidAt
		}
	}

	return cmd
}

func invoiceDescription(invoice *stripe.Invoice) string {
	if invoice.Lines == nil || len(invoice.Lines.Data) == 0 {
		return ""
	}
	return invoice.Lines.Data[0].Description
}

func subscriptionIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice.Parent != nil &&
		invoice.Parent.SubscriptionDetails != nil &&
		invoice.Parent.SubscriptionDetails.Subscription != nil &&
		invoice.Parent.SubscriptionDetails.Subscription.ID != "" {
		return invoice.Parent.SubscriptionDetails.Subscription.ID
	}

	if invoice.Lines != nil {
		for _, line := range invoice.Lines.Data {
			if line == nil || line.Parent == nil || line.Parent.SubscriptionItemDetails == nil {
				continue
			}
			if line.Parent.SubscriptionItemDetails.Subscription != "" {
				return line.Parent.SubscriptionItemDetails.Subscription
			}
		}
	}

	return ""
}

func customerIDFromInvoice(invoice *stripe.Invoice) string {
	if invoice.Customer != nil && invoice.Customer.ID != "" {
		return invoice.Customer.ID
	}
	return ""
}

func unixToTime(ts int64) time.Time {
	if ts <= 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0).UTC()
}
