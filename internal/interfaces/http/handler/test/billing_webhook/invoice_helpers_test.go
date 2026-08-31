package billingwebhooktest

import (
	"testing"
	"time"

	"go-api/internal/interfaces/http/handler"

	"github.com/stripe/stripe-go/v82"
)

func TestBillingWebhookInvoiceHelpers_SubscriptionIDFromParent(t *testing.T) {
	invoice := &stripe.Invoice{
		Parent: &stripe.InvoiceParent{
			SubscriptionDetails: &stripe.InvoiceParentSubscriptionDetails{
				Subscription: &stripe.Subscription{ID: "sub_parent"},
			},
		},
	}
	if got := handler.SubscriptionIDFromInvoiceForTest(invoice); got != "sub_parent" {
		t.Fatalf("subscription id: got %q want sub_parent", got)
	}
}

func TestBillingWebhookInvoiceHelpers_SubscriptionIDFromLines(t *testing.T) {
	invoice := &stripe.Invoice{
		Lines: &stripe.InvoiceLineItemList{
			Data: []*stripe.InvoiceLineItem{
				{Parent: nil},
				{
					Parent: &stripe.InvoiceLineItemParent{
						SubscriptionItemDetails: &stripe.InvoiceLineItemParentSubscriptionItemDetails{
							Subscription: "sub_line",
						},
					},
				},
			},
		},
	}
	if got := handler.SubscriptionIDFromInvoiceForTest(invoice); got != "sub_line" {
		t.Fatalf("subscription id: got %q want sub_line", got)
	}
}

func TestBillingWebhookInvoiceHelpers_SubscriptionIDEmpty(t *testing.T) {
	invoice := &stripe.Invoice{}
	if got := handler.SubscriptionIDFromInvoiceForTest(invoice); got != "" {
		t.Fatalf("subscription id: got %q want empty", got)
	}
}

func TestBillingWebhookInvoiceHelpers_CustomerID(t *testing.T) {
	withCustomer := &stripe.Invoice{Customer: &stripe.Customer{ID: "cus_123"}}
	if got := handler.CustomerIDFromInvoiceForTest(withCustomer); got != "cus_123" {
		t.Fatalf("customer id: got %q", got)
	}

	withoutCustomer := &stripe.Invoice{}
	if got := handler.CustomerIDFromInvoiceForTest(withoutCustomer); got != "" {
		t.Fatalf("customer id: got %q want empty", got)
	}
}

func TestBillingWebhookInvoiceHelpers_InvoiceDescription(t *testing.T) {
	empty := &stripe.Invoice{}
	if got := handler.InvoiceDescriptionForTest(empty); got != "" {
		t.Fatalf("description: got %q want empty", got)
	}

	withLines := &stripe.Invoice{
		Lines: &stripe.InvoiceLineItemList{
			Data: []*stripe.InvoiceLineItem{
				{Description: "Premium plan"},
			},
		},
	}
	if got := handler.InvoiceDescriptionForTest(withLines); got != "Premium plan" {
		t.Fatalf("description: got %q", got)
	}
}

func TestBillingWebhookInvoiceHelpers_UnixToTime(t *testing.T) {
	if got := handler.UnixToTimeForTest(0); !got.IsZero() {
		t.Fatalf("expected zero time for ts=0")
	}
	if got := handler.UnixToTimeForTest(-1); !got.IsZero() {
		t.Fatalf("expected zero time for negative ts")
	}

	want := time.Unix(1704067200, 0).UTC()
	if got := handler.UnixToTimeForTest(1704067200); !got.Equal(want) {
		t.Fatalf("time: got %v want %v", got, want)
	}
}

func TestBillingWebhookInvoiceHelpers_UpsertCommand(t *testing.T) {
	paidAt := int64(1704067200)
	invoice := &stripe.Invoice{
		ID:       "in_123",
		Customer: &stripe.Customer{ID: "cus_123"},
		Lines: &stripe.InvoiceLineItemList{
			Data: []*stripe.InvoiceLineItem{
				{
					Description: "Monthly subscription",
					Parent: &stripe.InvoiceLineItemParent{
						SubscriptionItemDetails: &stripe.InvoiceLineItemParentSubscriptionItemDetails{
							Subscription: "sub_line",
						},
					},
				},
			},
		},
		Number:         "INV-001",
		Status:         stripe.InvoiceStatusPaid,
		Currency:       stripe.CurrencyEUR,
		AmountDue:      1000,
		AmountPaid:     1000,
		Total:          1000,
		BillingReason:  stripe.InvoiceBillingReasonSubscriptionCycle,
		PeriodStart:    1704067200,
		PeriodEnd:      1706745600,
		Created:        1704067200,
		StatusTransitions: &stripe.InvoiceStatusTransitions{
			PaidAt: paidAt,
		},
	}

	cmd := handler.UpsertInvoiceCommandFromStripeForTest(invoice)
	if cmd.StripeInvoiceID != "in_123" {
		t.Fatalf("invoice id: got %q", cmd.StripeInvoiceID)
	}
	if cmd.StripeCustomerID != "cus_123" {
		t.Fatalf("customer id: got %q", cmd.StripeCustomerID)
	}
	if cmd.StripeSubscriptionID != "sub_line" {
		t.Fatalf("subscription id: got %q", cmd.StripeSubscriptionID)
	}
	if cmd.Description != "Monthly subscription" {
		t.Fatalf("description: got %q", cmd.Description)
	}
	if cmd.PaidAt == nil {
		t.Fatal("expected paid at to be set")
	}
	if !cmd.PaidAt.Equal(time.Unix(paidAt, 0).UTC()) {
		t.Fatalf("paid at: got %v", cmd.PaidAt)
	}
}

func TestBillingWebhookInvoiceHelpers_UpsertCommandWithoutPaidAt(t *testing.T) {
	invoice := &stripe.Invoice{
		ID:          "in_456",
		PeriodStart: 0,
		PeriodEnd:   0,
		Created:     0,
	}
	cmd := handler.UpsertInvoiceCommandFromStripeForTest(invoice)
	if cmd.PaidAt != nil {
		t.Fatalf("paid at: got %v want nil", cmd.PaidAt)
	}
	if !cmd.PeriodStart.IsZero() || !cmd.PeriodEnd.IsZero() || !cmd.StripeCreatedAt.IsZero() {
		t.Fatal("expected zero times for non-positive unix timestamps")
	}
}
