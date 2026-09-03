package billingwebhooktest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cmdsubscription "go-api/internal/application/command/subscription"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/stripe/stripe-go/v82"
)

type mockCheckoutCompletedHandler struct {
	called bool
	cmd    cmdsubscription.CheckoutCompletedCommand
	err    error
}

func (m *mockCheckoutCompletedHandler) Handle(_ context.Context, cmd cmdsubscription.CheckoutCompletedCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockSubscriptionUpdatedHandler struct {
	called bool
	cmd    cmdsubscription.SubscriptionUpdatedCommand
	err    error
}

func (m *mockSubscriptionUpdatedHandler) Handle(_ context.Context, cmd cmdsubscription.SubscriptionUpdatedCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockSubscriptionDeletedHandler struct {
	called bool
	cmd    cmdsubscription.SubscriptionDeletedCommand
	err    error
}

func (m *mockSubscriptionDeletedHandler) Handle(_ context.Context, cmd cmdsubscription.SubscriptionDeletedCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockInvoicePaymentSucceededHandler struct {
	called bool
	cmd    cmdsubscription.InvoicePaymentSucceededCommand
	err    error
}

func (m *mockInvoicePaymentSucceededHandler) Handle(_ context.Context, cmd cmdsubscription.InvoicePaymentSucceededCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockInvoicePaymentFailedHandler struct {
	called bool
	cmd    cmdsubscription.InvoicePaymentFailedCommand
	err    error
}

func (m *mockInvoicePaymentFailedHandler) Handle(_ context.Context, cmd cmdsubscription.InvoicePaymentFailedCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockUpsertInvoiceHandler struct {
	called bool
	cmd    cmdsubscription.UpsertInvoiceCommand
	err    error
}

func (m *mockUpsertInvoiceHandler) Handle(_ context.Context, cmd cmdsubscription.UpsertInvoiceCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type billingWebhookMocks struct {
	checkoutCompleted       *mockCheckoutCompletedHandler
	subscriptionUpdated     *mockSubscriptionUpdatedHandler
	subscriptionDeleted     *mockSubscriptionDeletedHandler
	invoicePaymentSucceeded *mockInvoicePaymentSucceededHandler
	invoicePaymentFailed    *mockInvoicePaymentFailedHandler
	upsertInvoice           *mockUpsertInvoiceHandler
}

func newBillingWebhookHandler(mocks billingWebhookMocks) *handler.BillingWebhookHandler {
	if mocks.checkoutCompleted == nil {
		mocks.checkoutCompleted = &mockCheckoutCompletedHandler{}
	}
	if mocks.subscriptionUpdated == nil {
		mocks.subscriptionUpdated = &mockSubscriptionUpdatedHandler{}
	}
	if mocks.subscriptionDeleted == nil {
		mocks.subscriptionDeleted = &mockSubscriptionDeletedHandler{}
	}
	if mocks.invoicePaymentSucceeded == nil {
		mocks.invoicePaymentSucceeded = &mockInvoicePaymentSucceededHandler{}
	}
	if mocks.invoicePaymentFailed == nil {
		mocks.invoicePaymentFailed = &mockInvoicePaymentFailedHandler{}
	}
	if mocks.upsertInvoice == nil {
		mocks.upsertInvoice = &mockUpsertInvoiceHandler{}
	}

	return handler.NewBillingWebhookHandler(
		mocks.checkoutCompleted,
		mocks.subscriptionUpdated,
		mocks.subscriptionDeleted,
		mocks.invoicePaymentSucceeded,
		mocks.invoicePaymentFailed,
		mocks.upsertInvoice,
	)
}

func executeWebhook(t *testing.T, h *handler.BillingWebhookHandler, event stripe.Event) *http.Response {
	t.Helper()

	app := testutil.NewTestApp()
	app.Post("/webhooks/billing", testutil.WithLocal("payload", event), h.Execute)

	req, err := testutil.JSONRequest(http.MethodPost, "/webhooks/billing", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	return resp
}

func stripeEvent(eventType string, data any) stripe.Event {
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return stripe.Event{
		ID:   "evt_test",
		Type: stripe.EventType(eventType),
		Data: &stripe.EventData{Raw: raw},
	}
}

func TestBillingWebhookHandler_Execute_CheckoutCompleted_Success(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{checkoutCompleted: checkout})

	event := stripeEvent("checkout.session.completed", map[string]any{
		"client_reference_id": testutil.TestUserID.String(),
		"customer":            map[string]string{"id": "cus_123"},
		"subscription":        map[string]string{"id": "sub_123"},
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !checkout.called {
		t.Fatal("expected checkout completed handler to be called")
	}
	if checkout.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", checkout.cmd.UserID)
	}
	if checkout.cmd.StripeCustomerID != "cus_123" {
		t.Fatalf("customer id: got %q", checkout.cmd.StripeCustomerID)
	}
	if checkout.cmd.StripeSubscriptionID != "sub_123" {
		t.Fatalf("subscription id: got %q", checkout.cmd.StripeSubscriptionID)
	}
}

func TestBillingWebhookHandler_Execute_SubscriptionUpdated_Success(t *testing.T) {
	updated := &mockSubscriptionUpdatedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{subscriptionUpdated: updated})

	event := stripeEvent("customer.subscription.updated", map[string]any{
		"id":                   "sub_123",
		"customer":             map[string]string{"id": "cus_123"},
		"status":               "active",
		"cancel_at_period_end": false,
		"items": map[string]any{
			"data": []map[string]any{
				{
					"id":                     "si_123",
					"price":                  map[string]string{"id": "price_123"},
					"current_period_start":   1704067200,
					"current_period_end":     1706745600,
				},
			},
		},
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !updated.called {
		t.Fatal("expected subscription updated handler to be called")
	}
	if updated.cmd.StripeSubscriptionID != "sub_123" {
		t.Fatalf("subscription id: got %q", updated.cmd.StripeSubscriptionID)
	}
	if updated.cmd.StripeCustomerID != "cus_123" {
		t.Fatalf("customer id: got %q", updated.cmd.StripeCustomerID)
	}
	if updated.cmd.StripePriceID != "price_123" {
		t.Fatalf("price id: got %q", updated.cmd.StripePriceID)
	}
}

func TestBillingWebhookHandler_Execute_SubscriptionDeleted_Success(t *testing.T) {
	deleted := &mockSubscriptionDeletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{subscriptionDeleted: deleted})

	event := stripeEvent("customer.subscription.deleted", map[string]any{
		"id": "sub_123",
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !deleted.called {
		t.Fatal("expected subscription deleted handler to be called")
	}
	if deleted.cmd.StripeSubscriptionID != "sub_123" {
		t.Fatalf("subscription id: got %q", deleted.cmd.StripeSubscriptionID)
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentSucceeded_Success(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{}
	succeeded := &mockInvoicePaymentSucceededHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{
		upsertInvoice:           upsert,
		invoicePaymentSucceeded: succeeded,
	})

	event := stripeEvent("invoice.payment_succeeded", sampleInvoicePayload())

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !upsert.called {
		t.Fatal("expected upsert invoice handler to be called")
	}
	if !succeeded.called {
		t.Fatal("expected invoice payment succeeded handler to be called")
	}
	if upsert.cmd.StripeInvoiceID != "in_123" {
		t.Fatalf("invoice id: got %q", upsert.cmd.StripeInvoiceID)
	}
	if succeeded.cmd.StripeSubscriptionID != "sub_123" {
		t.Fatalf("subscription id: got %q", succeeded.cmd.StripeSubscriptionID)
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentFailed_Success(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{}
	failed := &mockInvoicePaymentFailedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{
		upsertInvoice:        upsert,
		invoicePaymentFailed: failed,
	})

	event := stripeEvent("invoice.payment_failed", sampleInvoicePayload())

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !upsert.called {
		t.Fatal("expected upsert invoice handler to be called")
	}
	if !failed.called {
		t.Fatal("expected invoice payment failed handler to be called")
	}
	if failed.cmd.StripeSubscriptionID != "sub_123" {
		t.Fatalf("subscription id: got %q", failed.cmd.StripeSubscriptionID)
	}
}

func TestBillingWebhookHandler_Execute_UnknownEventType_Success(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{}
	updated := &mockSubscriptionUpdatedHandler{}
	deleted := &mockSubscriptionDeletedHandler{}
	upsert := &mockUpsertInvoiceHandler{}
	succeeded := &mockInvoicePaymentSucceededHandler{}
	failed := &mockInvoicePaymentFailedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{
		checkoutCompleted:       checkout,
		subscriptionUpdated:     updated,
		subscriptionDeleted:     deleted,
		upsertInvoice:           upsert,
		invoicePaymentSucceeded: succeeded,
		invoicePaymentFailed:    failed,
	})

	event := stripeEvent("customer.created", map[string]any{"id": "cus_999"})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if checkout.called || updated.called || deleted.called || upsert.called || succeeded.called || failed.called {
		t.Fatal("no handler should be called for unknown event type")
	}
}

func TestBillingWebhookHandler_Execute_HandlerError_Internal(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{err: errors.New("database unavailable")}
	h := newBillingWebhookHandler(billingWebhookMocks{checkoutCompleted: checkout})

	event := stripeEvent("checkout.session.completed", map[string]any{
		"client_reference_id": testutil.TestUserID.String(),
		"customer":            map[string]string{"id": "cus_123"},
		"subscription":        map[string]string{"id": "sub_123"},
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "failed to process event" {
		t.Fatalf("error message: got %v", body["message"])
	}
}

func TestBillingWebhookHandler_Execute_HandlerError_SubscriptionNotLinked(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{err: cmdsubscription.ErrStripeSubscriptionNotLinked}
	h := newBillingWebhookHandler(billingWebhookMocks{upsertInvoice: upsert})

	event := stripeEvent("invoice.payment_succeeded", sampleInvoicePayload())

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != fiber.StatusConflict {
		t.Fatalf("status: got %d want %d", resp.StatusCode, fiber.StatusConflict)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "subscription not linked yet" {
		t.Fatalf("error message: got %v", body["message"])
	}
}

func TestBillingWebhookHandler_Execute_CheckoutCompleted_InvalidUserID(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{checkoutCompleted: checkout})

	event := stripeEvent("checkout.session.completed", map[string]any{
		"client_reference_id": "not-a-uuid",
		"customer":            map[string]string{"id": "cus_123"},
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if checkout.called {
		t.Fatal("checkout handler must not be called with invalid user id")
	}
}

func TestBillingWebhookHandler_Execute_CheckoutCompleted_WithoutCustomerOrSubscription(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{checkoutCompleted: checkout})

	event := stripeEvent("checkout.session.completed", map[string]any{
		"client_reference_id": testutil.TestUserID.String(),
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !checkout.called {
		t.Fatal("expected checkout completed handler to be called")
	}
	if checkout.cmd.StripeCustomerID != "" || checkout.cmd.StripeSubscriptionID != "" {
		t.Fatalf("expected empty customer/subscription ids, got customer=%q subscription=%q",
			checkout.cmd.StripeCustomerID, checkout.cmd.StripeSubscriptionID)
	}
}

func TestBillingWebhookHandler_Execute_CheckoutCompleted_CustomerOnly(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{checkoutCompleted: checkout})

	event := stripeEvent("checkout.session.completed", map[string]any{
		"client_reference_id": testutil.TestUserID.String(),
		"customer":            map[string]string{"id": "cus_123"},
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if checkout.cmd.StripeCustomerID != "cus_123" {
		t.Fatalf("customer id: got %q", checkout.cmd.StripeCustomerID)
	}
	if checkout.cmd.StripeSubscriptionID != "" {
		t.Fatalf("subscription id: got %q want empty", checkout.cmd.StripeSubscriptionID)
	}
}

func TestBillingWebhookHandler_Execute_CheckoutCompleted_SubscriptionOnly(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{checkoutCompleted: checkout})

	event := stripeEvent("checkout.session.completed", map[string]any{
		"client_reference_id": testutil.TestUserID.String(),
		"subscription":        map[string]string{"id": "sub_123"},
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if checkout.cmd.StripeSubscriptionID != "sub_123" {
		t.Fatalf("subscription id: got %q", checkout.cmd.StripeSubscriptionID)
	}
	if checkout.cmd.StripeCustomerID != "" {
		t.Fatalf("customer id: got %q want empty", checkout.cmd.StripeCustomerID)
	}
}

func TestBillingWebhookHandler_Execute_SubscriptionUpdated_HandlerError(t *testing.T) {
	updated := &mockSubscriptionUpdatedHandler{err: errors.New("database unavailable")}
	h := newBillingWebhookHandler(billingWebhookMocks{subscriptionUpdated: updated})

	event := stripeEvent("customer.subscription.updated", map[string]any{
		"id":       "sub_123",
		"customer": map[string]string{"id": "cus_123"},
		"status":   "active",
		"items": map[string]any{
			"data": []map[string]any{
				{
					"price":                map[string]string{"id": "price_123"},
					"current_period_start": 1704067200,
					"current_period_end":   1706745600,
				},
			},
		},
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBillingWebhookHandler_Execute_SubscriptionDeleted_UnmarshalError(t *testing.T) {
	deleted := &mockSubscriptionDeletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{subscriptionDeleted: deleted})

	event := stripe.Event{
		ID:   "evt_test",
		Type: stripe.EventType("customer.subscription.deleted"),
		Data: &stripe.EventData{Raw: json.RawMessage(`{invalid`)},
	}

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if deleted.called {
		t.Fatal("deleted handler must not be called on unmarshal error")
	}
}

func TestBillingWebhookHandler_Execute_SubscriptionDeleted_HandlerError(t *testing.T) {
	deleted := &mockSubscriptionDeletedHandler{err: errors.New("database unavailable")}
	h := newBillingWebhookHandler(billingWebhookMocks{subscriptionDeleted: deleted})

	event := stripeEvent("customer.subscription.deleted", map[string]any{"id": "sub_123"})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentSucceeded_UnmarshalError(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{upsertInvoice: upsert})

	event := stripe.Event{
		ID:   "evt_test",
		Type: stripe.EventType("invoice.payment_succeeded"),
		Data: &stripe.EventData{Raw: json.RawMessage(`{invalid`)},
	}

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if upsert.called {
		t.Fatal("upsert handler must not be called on unmarshal error")
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentSucceeded_UpsertError(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{err: errors.New("database unavailable")}
	h := newBillingWebhookHandler(billingWebhookMocks{upsertInvoice: upsert})

	event := stripeEvent("invoice.payment_succeeded", sampleInvoicePayload())

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentSucceeded_HandlerError(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{}
	succeeded := &mockInvoicePaymentSucceededHandler{err: errors.New("database unavailable")}
	h := newBillingWebhookHandler(billingWebhookMocks{
		upsertInvoice:           upsert,
		invoicePaymentSucceeded: succeeded,
	})

	event := stripeEvent("invoice.payment_succeeded", sampleInvoicePayload())

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentFailed_UpsertError(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{err: errors.New("database unavailable")}
	h := newBillingWebhookHandler(billingWebhookMocks{upsertInvoice: upsert})

	event := stripeEvent("invoice.payment_failed", sampleInvoicePayload())

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentFailed_HandlerError(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{}
	failed := &mockInvoicePaymentFailedHandler{err: errors.New("database unavailable")}
	h := newBillingWebhookHandler(billingWebhookMocks{
		upsertInvoice:        upsert,
		invoicePaymentFailed: failed,
	})

	event := stripeEvent("invoice.payment_failed", sampleInvoicePayload())

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBillingWebhookHandler_Execute_SubscriptionUpdated_UnmarshalError(t *testing.T) {
	updated := &mockSubscriptionUpdatedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{subscriptionUpdated: updated})

	event := stripe.Event{
		ID:   "evt_test",
		Type: stripe.EventType("customer.subscription.updated"),
		Data: &stripe.EventData{Raw: json.RawMessage(`{invalid`)},
	}

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestBillingWebhookHandler_Execute_CheckoutCompleted_StringCustomerAndSubscription(t *testing.T) {
	checkout := &mockCheckoutCompletedHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{checkoutCompleted: checkout})

	event := stripeEvent("checkout.session.completed", map[string]any{
		"client_reference_id": testutil.TestUserID.String(),
		"customer":            "cus_string",
		"subscription":        "sub_string",
	})

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !checkout.called {
		t.Fatal("expected checkout completed handler to be called")
	}
}

func TestBillingWebhookHandler_Execute_InvoicePaymentFailed_UnmarshalError(t *testing.T) {
	upsert := &mockUpsertInvoiceHandler{}
	h := newBillingWebhookHandler(billingWebhookMocks{upsertInvoice: upsert})

	event := stripe.Event{
		ID:   "evt_test",
		Type: stripe.EventType("invoice.payment_failed"),
		Data: &stripe.EventData{Raw: json.RawMessage(`{invalid`)},
	}

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
	if upsert.called {
		t.Fatal("upsert handler must not be called on unmarshal error")
	}
}

func sampleInvoicePayload() map[string]any {
	return map[string]any{
		"id":       "in_123",
		"customer": map[string]string{"id": "cus_123"},
		"parent": map[string]any{
			"subscription_details": map[string]any{
				"subscription": map[string]string{"id": "sub_123"},
			},
		},
		"number":         "INV-001",
		"status":         "paid",
		"currency":       "eur",
		"amount_due":     1000,
		"amount_paid":    1000,
		"total":          1000,
		"billing_reason": "subscription_cycle",
		"period_start":   1704067200,
		"period_end":     1706745600,
		"created":        1704067200,
	}
}
