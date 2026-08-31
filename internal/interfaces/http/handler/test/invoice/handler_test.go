package invoicetest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	queryinvoice "go-api/internal/application/query/invoice"
	domaininvoice "go-api/internal/domain/invoice"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

var testInvoiceID = uuid.MustParse("01960000-0000-7000-8000-00000000000b")

type mockListInvoicesHandler struct {
	called bool
	query  queryinvoice.ListInvoicesQuery
	views  []*domaininvoice.InvoiceView
	total  int64
	err    error
}

func (m *mockListInvoicesHandler) Handle(
	_ context.Context,
	q queryinvoice.ListInvoicesQuery,
) ([]*domaininvoice.InvoiceView, int64, error) {
	m.called = true
	m.query = q
	return m.views, m.total, m.err
}

func newInvoiceHandler(list *mockListInvoicesHandler) *handler.InvoiceHandler {
	if list == nil {
		list = &mockListInvoicesHandler{}
	}
	return handler.NewInvoiceHandler(list)
}

func sampleInvoiceView() *domaininvoice.InvoiceView {
	return &domaininvoice.InvoiceView{
		ID:              testInvoiceID,
		UserID:          testutil.TestUserID,
		StripeInvoiceID: "in_123",
		Status:          "paid",
		Currency:        "eur",
		AmountDue:       2999,
		AmountPaid:      2999,
		Total:           2999,
		StripeCreatedAt: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		CreatedAt:       time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

func TestInvoiceHandler_GetInvoices_Success(t *testing.T) {
	list := &mockListInvoicesHandler{
		views: []*domaininvoice.InvoiceView{sampleInvoiceView()},
		total: 1,
	}
	h := newInvoiceHandler(list)

	app := testutil.NewTestApp()
	app.Get("/invoices", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetInvoices)

	req, err := testutil.JSONRequest(http.MethodGet, "/invoices?page=1&limit=10", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !list.called {
		t.Fatal("expected list handler to be called")
	}
	if list.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", list.query.UserID)
	}
	if list.query.Query.SortBy != "stripe_created_at" {
		t.Fatalf("sort by: got %q want stripe_created_at", list.query.Query.SortBy)
	}

	var out struct {
		Members []presenter.InvoiceResponse `json:"members"`
		Total   int                        `json:"total"`
	}
	testutil.DecodeJSON(t, resp, &out)
	if len(out.Members) != 1 {
		t.Fatalf("members length: got %d want 1", len(out.Members))
	}
	if out.Total != 1 {
		t.Fatalf("total: got %d want 1", out.Total)
	}
}

func TestInvoiceHandler_GetInvoices_Unauthorized(t *testing.T) {
	list := &mockListInvoicesHandler{}
	h := newInvoiceHandler(list)

	app := testutil.NewTestApp()
	app.Get("/invoices", h.GetInvoices)

	req, err := testutil.JSONRequest(http.MethodGet, "/invoices", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if list.called {
		t.Fatal("list handler must not be called without auth")
	}
}

func TestInvoiceHandler_GetInvoices_InvalidQuery(t *testing.T) {
	list := &mockListInvoicesHandler{}
	h := newInvoiceHandler(list)

	app := testutil.NewTestApp()
	app.Get("/invoices", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetInvoices)

	req, err := testutil.JSONRequest(http.MethodGet, "/invoices?page=not-a-number", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if list.called {
		t.Fatal("list handler must not be called with invalid query")
	}
}

func TestInvoiceHandler_GetInvoices_HandlerError_Internal(t *testing.T) {
	list := &mockListInvoicesHandler{err: errors.New("database unavailable")}
	h := newInvoiceHandler(list)

	app := testutil.NewTestApp()
	app.Get("/invoices", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetInvoices)

	req, err := testutil.JSONRequest(http.MethodGet, "/invoices", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
