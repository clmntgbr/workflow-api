package plantest

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	queryplan "go-api/internal/application/query/plan"
	domainplan "go-api/internal/domain/plan"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"
)

type mockListActivePlansHandler struct {
	called bool
	views  []domainplan.PlanView
	err    error
}

func (m *mockListActivePlansHandler) Handle(
	_ context.Context,
	_ queryplan.ListActivePlansQuery,
) ([]domainplan.PlanView, error) {
	m.called = true
	return m.views, m.err
}

func newPlanHandler(list *mockListActivePlansHandler) *handler.PlanHandler {
	if list == nil {
		list = &mockListActivePlansHandler{}
	}
	return handler.NewPlanHandler(list)
}

func samplePlanView() domainplan.PlanView {
	return domainplan.PlanView{
		ID:              testutil.TestPlanID,
		Name:            "Pro",
		Description:     "Professional plan",
		Slug:            "pro",
		StripePriceID:   "price_123",
		IsActive:        true,
		BillingInterval: domainplan.BillingIntervalMonth,
		Price:           29.99,
		Currency:        domainplan.CurrencyEUR,
		CreatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestPlanHandler_List_Success(t *testing.T) {
	list := &mockListActivePlansHandler{views: []domainplan.PlanView{samplePlanView()}}
	h := newPlanHandler(list)

	app := testutil.NewTestApp()
	app.Get("/plans", h.List)

	req, err := testutil.JSONRequest(http.MethodGet, "/plans", nil)
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

	var out []presenter.PlanResponse
	testutil.DecodeJSON(t, resp, &out)
	if len(out) != 1 {
		t.Fatalf("plans length: got %d want 1", len(out))
	}
	if out[0].ID != testutil.TestPlanID.String() {
		t.Fatalf("plan id: got %s", out[0].ID)
	}
}

func TestPlanHandler_List_HandlerError_Internal(t *testing.T) {
	list := &mockListActivePlansHandler{err: errors.New("database unavailable")}
	h := newPlanHandler(list)

	app := testutil.NewTestApp()
	app.Get("/plans", h.List)

	req, err := testutil.JSONRequest(http.MethodGet, "/plans", nil)
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
