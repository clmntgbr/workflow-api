package subscriptiontest

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	cmdsubscription "go-api/internal/application/command/subscription"
	querysubscription "go-api/internal/application/query/subscription"
	"go-api/internal/domain/port"
	domainplan "go-api/internal/domain/plan"
	domainsubscription "go-api/internal/domain/subscription"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/presenter"
	"go-api/internal/interfaces/http/testutil"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

var testSubscriptionID = uuid.MustParse("01960000-0000-7000-8000-00000000000d")

type mockGetCurrentSubscriptionHandler struct {
	called bool
	query  querysubscription.GetCurrentSubscriptionQuery
	view   *domainsubscription.SubscriptionView
	err    error
}

func (m *mockGetCurrentSubscriptionHandler) Handle(
	_ context.Context,
	q querysubscription.GetCurrentSubscriptionQuery,
) (*domainsubscription.SubscriptionView, error) {
	m.called = true
	m.query = q
	return m.view, m.err
}

type mockGetQuotaUsageHandler struct {
	called bool
	query  querysubscription.GetQuotaUsageQuery
	usage  *querysubscription.QuotaUsageView
	err    error
}

func (m *mockGetQuotaUsageHandler) Handle(
	_ context.Context,
	q querysubscription.GetQuotaUsageQuery,
) (*querysubscription.QuotaUsageView, error) {
	m.called = true
	m.query = q
	return m.usage, m.err
}

type mockPreviewPlanChangeHandler struct {
	called  bool
	query   querysubscription.PreviewPlanChangeQuery
	preview *querysubscription.PlanChangePreview
	err     error
}

func (m *mockPreviewPlanChangeHandler) Handle(
	_ context.Context,
	q querysubscription.PreviewPlanChangeQuery,
) (*querysubscription.PlanChangePreview, error) {
	m.called = true
	m.query = q
	return m.preview, m.err
}

type mockCreateSubscriptionHandler struct {
	called bool
	cmd    cmdsubscription.CreateSubscriptionCommand
	result *cmdsubscription.CreateSubscriptionResult
	err    error
}

func (m *mockCreateSubscriptionHandler) Handle(
	_ context.Context,
	cmd cmdsubscription.CreateSubscriptionCommand,
) (*cmdsubscription.CreateSubscriptionResult, error) {
	m.called = true
	m.cmd = cmd
	return m.result, m.err
}

type mockCreateBillingPortalHandler struct {
	called bool
	cmd    cmdsubscription.CreateBillingPortalCommand
	url    string
	err    error
}

func (m *mockCreateBillingPortalHandler) Handle(
	_ context.Context,
	cmd cmdsubscription.CreateBillingPortalCommand,
) (string, error) {
	m.called = true
	m.cmd = cmd
	return m.url, m.err
}

func newSubscriptionHandler(
	getCurrent *mockGetCurrentSubscriptionHandler,
	getQuota *mockGetQuotaUsageHandler,
	preview *mockPreviewPlanChangeHandler,
	create *mockCreateSubscriptionHandler,
	billingPortal *mockCreateBillingPortalHandler,
) *handler.SubscriptionHandler {
	if getCurrent == nil {
		getCurrent = &mockGetCurrentSubscriptionHandler{}
	}
	if getQuota == nil {
		getQuota = &mockGetQuotaUsageHandler{}
	}
	if preview == nil {
		preview = &mockPreviewPlanChangeHandler{}
	}
	if create == nil {
		create = &mockCreateSubscriptionHandler{}
	}
	if billingPortal == nil {
		billingPortal = &mockCreateBillingPortalHandler{}
	}
	return handler.NewSubscriptionHandler(getCurrent, getQuota, preview, create, billingPortal)
}

func activeProject() fiber.Handler {
	return testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID)
}

func sampleSubscriptionView() *domainsubscription.SubscriptionView {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := now.AddDate(0, 1, 0)
	return &domainsubscription.SubscriptionView{
		ID:                   testSubscriptionID,
		PlanID:               testutil.TestPlanID,
		StripeCustomerID:     "cus_123",
		StripeSubscriptionID: "sub_123",
		Status:               domainsubscription.StatusActive,
		StartDate:            now,
		EndDate:              end,
		QuotaPeriodStart:     now,
		Plan: &domainplan.PlanView{
			ID:              testutil.TestPlanID,
			Name:            "Pro",
			Slug:            "pro",
			StripePriceID:   "price_123",
			IsActive:        true,
			BillingInterval: domainplan.BillingIntervalMonth,
			Price:           29.99,
			Currency:        domainplan.CurrencyEUR,
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func sampleQuotaUsageView() *querysubscription.QuotaUsageView {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &querysubscription.QuotaUsageView{
		WorkflowRuns: querysubscription.MonthlyQuotaCounter{
			PeriodStart: now,
			PeriodEnd:   now.AddDate(0, 1, 0),
			Used:        5,
			Max:         100,
			Left:        95,
		},
		Projects: querysubscription.QuotaCounter{Used: 1, Max: 3, Left: 2},
		Limits: querysubscription.QuotaLimits{
			MaxStepsPerWorkflow:      50,
			MaxVariablesPerWorkflow:  100,
			MaxAssertionsPerWorkflow: 100,
			AllowsInsights:             true,
		},
	}
}

func samplePlanChangePreview() *querysubscription.PlanChangePreview {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return &querysubscription.PlanChangePreview{
		RequiresCheckout: false,
		Currency:         "eur",
		AmountDue:        999,
		Subtotal:         999,
		Total:            999,
		PeriodStart:      now,
		PeriodEnd:        now.AddDate(0, 1, 0),
		TargetPlanID:     testutil.TestPlanID.String(),
		TargetPlanSlug:   "pro",
		TargetPlanName:   "Pro",
		TargetPlanPrice:  29.99,
		Lines:            []port.ProrationPreviewLine{},
	}
}

func validPreviewBody() map[string]any {
	return map[string]any{"planId": testutil.TestPlanID.String()}
}

func validCreateBody() map[string]any {
	return map[string]any{"planId": testutil.TestPlanID.String()}
}

func mustJSONRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()
	req, err := testutil.JSONRequest(method, path, body)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	return req
}

func TestSubscriptionHandler_GetSubscription_Success(t *testing.T) {
	getCurrent := &mockGetCurrentSubscriptionHandler{view: sampleSubscriptionView()}
	h := newSubscriptionHandler(getCurrent, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription", activeProject(), h.GetSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !getCurrent.called {
		t.Fatal("expected get current subscription handler to be called")
	}
	if getCurrent.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", getCurrent.query.UserID)
	}

	var out presenter.SubscriptionResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.ID != testSubscriptionID.String() {
		t.Fatalf("subscription id: got %s", out.ID)
	}
}

func TestSubscriptionHandler_GetSubscription_Unauthorized(t *testing.T) {
	getCurrent := &mockGetCurrentSubscriptionHandler{}
	h := newSubscriptionHandler(getCurrent, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription", h.GetSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if getCurrent.called {
		t.Fatal("get current handler must not be called without user")
	}
}

func TestSubscriptionHandler_GetSubscription_NotFound(t *testing.T) {
	getCurrent := &mockGetCurrentSubscriptionHandler{err: querysubscription.ErrSubscriptionNotFound}
	h := newSubscriptionHandler(getCurrent, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription", activeProject(), h.GetSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSubscriptionHandler_GetSubscription_HandlerError_Internal(t *testing.T) {
	getCurrent := &mockGetCurrentSubscriptionHandler{err: errors.New("database unavailable")}
	h := newSubscriptionHandler(getCurrent, nil, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription", activeProject(), h.GetSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestSubscriptionHandler_GetQuota_Success(t *testing.T) {
	getQuota := &mockGetQuotaUsageHandler{usage: sampleQuotaUsageView()}
	h := newSubscriptionHandler(nil, getQuota, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription/quota", activeProject(), h.GetQuota)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription/quota", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !getQuota.called {
		t.Fatal("expected get quota handler to be called")
	}
	if getQuota.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", getQuota.query.UserID)
	}
	if getQuota.query.ProjectID != testutil.TestProjectID {
		t.Fatalf("project id: got %s", getQuota.query.ProjectID)
	}

	var out presenter.QuotaUsageResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.WorkflowRuns.Used != 5 {
		t.Fatalf("workflow runs used: got %d want 5", out.WorkflowRuns.Used)
	}
	if out.Limits.MaxVariablesPerWorkflow != 100 {
		t.Fatalf("max variables: got %d want 100", out.Limits.MaxVariablesPerWorkflow)
	}
	if out.Limits.MaxAssertionsPerWorkflow != 100 {
		t.Fatalf("max assertions: got %d want 100", out.Limits.MaxAssertionsPerWorkflow)
	}
}

func TestSubscriptionHandler_GetQuota_Unauthorized(t *testing.T) {
	getQuota := &mockGetQuotaUsageHandler{}
	h := newSubscriptionHandler(nil, getQuota, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription/quota", h.GetQuota)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription/quota", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if getQuota.called {
		t.Fatal("get quota handler must not be called without user")
	}
}

func TestSubscriptionHandler_GetQuota_MissingActiveProject(t *testing.T) {
	getQuota := &mockGetQuotaUsageHandler{}
	h := newSubscriptionHandler(nil, getQuota, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription/quota", testutil.WithUserWithoutProject(testutil.TestUserID), h.GetQuota)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription/quota", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if getQuota.called {
		t.Fatal("get quota handler must not be called without active project")
	}
}

func TestSubscriptionHandler_GetQuota_NotFound(t *testing.T) {
	getQuota := &mockGetQuotaUsageHandler{err: querysubscription.ErrSubscriptionNotFound}
	h := newSubscriptionHandler(nil, getQuota, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription/quota", activeProject(), h.GetQuota)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription/quota", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSubscriptionHandler_GetQuota_ActiveProjectRequired(t *testing.T) {
	getQuota := &mockGetQuotaUsageHandler{err: querysubscription.ErrActiveProjectRequired}
	h := newSubscriptionHandler(nil, getQuota, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription/quota", activeProject(), h.GetQuota)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription/quota", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSubscriptionHandler_GetQuota_HandlerError_Internal(t *testing.T) {
	getQuota := &mockGetQuotaUsageHandler{err: errors.New("database unavailable")}
	h := newSubscriptionHandler(nil, getQuota, nil, nil, nil)

	app := testutil.NewTestApp()
	app.Get("/subscription/quota", activeProject(), h.GetQuota)

	resp, err := app.Test(mustJSONRequest(t, http.MethodGet, "/subscription/quota", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestSubscriptionHandler_PreviewSubscription_Success(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{preview: samplePlanChangePreview()}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", validPreviewBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !preview.called {
		t.Fatal("expected preview handler to be called")
	}
	if preview.query.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", preview.query.UserID)
	}
	if preview.query.PlanID != testutil.TestPlanID {
		t.Fatalf("plan id: got %s", preview.query.PlanID)
	}

	var out presenter.PlanChangePreviewResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.TargetPlanID != testutil.TestPlanID.String() {
		t.Fatalf("target plan id: got %s", out.TargetPlanID)
	}
}

func TestSubscriptionHandler_PreviewSubscription_Unauthorized(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", validPreviewBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if preview.called {
		t.Fatal("preview handler must not be called without user")
	}
}

func TestSubscriptionHandler_PreviewSubscription_InvalidRequestBody(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	req, err := http.NewRequest(http.MethodPost, "/subscription/preview", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if preview.called {
		t.Fatal("preview handler must not be called with invalid body")
	}
}

func TestSubscriptionHandler_PreviewSubscription_MissingPlanID(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", map[string]any{}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if preview.called {
		t.Fatal("preview handler must not be called without plan id")
	}
}

func TestSubscriptionHandler_PreviewSubscription_InvalidPlanID(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", map[string]any{"planId": "bad-id"}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if preview.called {
		t.Fatal("preview handler must not be called with invalid plan id")
	}
}

func TestSubscriptionHandler_PreviewSubscription_PlanNotFound(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{err: querysubscription.ErrPlanNotFound}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", validPreviewBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSubscriptionHandler_PreviewSubscription_PlanInactive(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{err: querysubscription.ErrPlanInactive}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", validPreviewBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSubscriptionHandler_PreviewSubscription_FreePlanCheckout(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{err: querysubscription.ErrFreePlanCheckout}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", validPreviewBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSubscriptionHandler_PreviewSubscription_AlreadyOnPlan(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{err: querysubscription.ErrAlreadyOnPlan}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", validPreviewBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSubscriptionHandler_PreviewSubscription_HandlerError_Internal(t *testing.T) {
	preview := &mockPreviewPlanChangeHandler{err: errors.New("stripe unavailable")}
	h := newSubscriptionHandler(nil, nil, preview, nil, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription/preview", activeProject(), h.PreviewSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/preview", validPreviewBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestSubscriptionHandler_CreateSubscription_Success(t *testing.T) {
	create := &mockCreateSubscriptionHandler{
		result: &cmdsubscription.CreateSubscriptionResult{URL: "https://checkout.stripe.com/session", Updated: false},
	}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", activeProject(), h.CreateSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription", validCreateBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !create.called {
		t.Fatal("expected create handler to be called")
	}
	if create.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", create.cmd.UserID)
	}
	if create.cmd.PlanID != testutil.TestPlanID {
		t.Fatalf("plan id: got %s", create.cmd.PlanID)
	}

	var out presenter.ChangeSubscriptionResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.URL == nil || *out.URL != "https://checkout.stripe.com/session" {
		t.Fatalf("url: got %v", out.URL)
	}
}

func TestSubscriptionHandler_CreateSubscription_Unauthorized(t *testing.T) {
	create := &mockCreateSubscriptionHandler{}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", h.CreateSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription", validCreateBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if create.called {
		t.Fatal("create handler must not be called without user")
	}
}

func TestSubscriptionHandler_CreateSubscription_InvalidRequestBody(t *testing.T) {
	create := &mockCreateSubscriptionHandler{}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", activeProject(), h.CreateSubscription)

	req, err := http.NewRequest(http.MethodPost, "/subscription", bytes.NewBufferString("not-json"))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid body")
	}
}

func TestSubscriptionHandler_CreateSubscription_InvalidPlanID(t *testing.T) {
	create := &mockCreateSubscriptionHandler{}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", activeProject(), h.CreateSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription", map[string]any{"planId": "bad-id"}))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
	if create.called {
		t.Fatal("create handler must not be called with invalid plan id")
	}
}

func TestSubscriptionHandler_CreateSubscription_PlanNotFound(t *testing.T) {
	create := &mockCreateSubscriptionHandler{err: querysubscription.ErrPlanNotFound}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", activeProject(), h.CreateSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription", validCreateBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSubscriptionHandler_CreateSubscription_SubscriptionNotFound(t *testing.T) {
	create := &mockCreateSubscriptionHandler{err: querysubscription.ErrSubscriptionNotFound}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", activeProject(), h.CreateSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription", validCreateBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSubscriptionHandler_CreateSubscription_AlreadyOnPlan(t *testing.T) {
	create := &mockCreateSubscriptionHandler{err: querysubscription.ErrAlreadyOnPlan}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", activeProject(), h.CreateSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription", validCreateBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSubscriptionHandler_CreateSubscription_HandlerError_Internal(t *testing.T) {
	create := &mockCreateSubscriptionHandler{err: errors.New("stripe unavailable")}
	h := newSubscriptionHandler(nil, nil, nil, create, nil)

	app := testutil.NewTestApp()
	app.Post("/subscription", activeProject(), h.CreateSubscription)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription", validCreateBody()))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}

func TestSubscriptionHandler_CreateBillingPortal_Success(t *testing.T) {
	billingPortal := &mockCreateBillingPortalHandler{url: "https://billing.stripe.com/portal"}
	h := newSubscriptionHandler(nil, nil, nil, nil, billingPortal)

	app := testutil.NewTestApp()
	app.Post("/subscription/billing-portal", activeProject(), h.CreateBillingPortal)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/billing-portal", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusOK)
	}
	if !billingPortal.called {
		t.Fatal("expected billing portal handler to be called")
	}
	if billingPortal.cmd.UserID != testutil.TestUserID {
		t.Fatalf("user id: got %s", billingPortal.cmd.UserID)
	}

	var out presenter.CheckoutSessionResponse
	testutil.DecodeJSON(t, resp, &out)
	if out.URL != "https://billing.stripe.com/portal" {
		t.Fatalf("url: got %s", out.URL)
	}
}

func TestSubscriptionHandler_CreateBillingPortal_Unauthorized(t *testing.T) {
	billingPortal := &mockCreateBillingPortalHandler{}
	h := newSubscriptionHandler(nil, nil, nil, nil, billingPortal)

	app := testutil.NewTestApp()
	app.Post("/subscription/billing-portal", h.CreateBillingPortal)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/billing-portal", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusUnauthorized)
	}
	if billingPortal.called {
		t.Fatal("billing portal handler must not be called without user")
	}
}

func TestSubscriptionHandler_CreateBillingPortal_NotFound(t *testing.T) {
	billingPortal := &mockCreateBillingPortalHandler{err: querysubscription.ErrSubscriptionNotFound}
	h := newSubscriptionHandler(nil, nil, nil, nil, billingPortal)

	app := testutil.NewTestApp()
	app.Post("/subscription/billing-portal", activeProject(), h.CreateBillingPortal)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/billing-portal", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestSubscriptionHandler_CreateBillingPortal_MissingStripeCustomer(t *testing.T) {
	billingPortal := &mockCreateBillingPortalHandler{err: cmdsubscription.ErrMissingStripeCustomer}
	h := newSubscriptionHandler(nil, nil, nil, nil, billingPortal)

	app := testutil.NewTestApp()
	app.Post("/subscription/billing-portal", activeProject(), h.CreateBillingPortal)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/billing-portal", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSubscriptionHandler_CreateBillingPortal_FreePlan(t *testing.T) {
	billingPortal := &mockCreateBillingPortalHandler{err: cmdsubscription.ErrFreePlanBillingPortal}
	h := newSubscriptionHandler(nil, nil, nil, nil, billingPortal)

	app := testutil.NewTestApp()
	app.Post("/subscription/billing-portal", activeProject(), h.CreateBillingPortal)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/billing-portal", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestSubscriptionHandler_CreateBillingPortal_HandlerError_Internal(t *testing.T) {
	billingPortal := &mockCreateBillingPortalHandler{err: errors.New("stripe unavailable")}
	h := newSubscriptionHandler(nil, nil, nil, nil, billingPortal)

	app := testutil.NewTestApp()
	app.Post("/subscription/billing-portal", activeProject(), h.CreateBillingPortal)

	resp, err := app.Test(mustJSONRequest(t, http.MethodPost, "/subscription/billing-portal", nil))
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}
}
