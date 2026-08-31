package userwebhooktest

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	usercmd "go-api/internal/application/command/user"
	domainuser "go-api/internal/domain/user"
	"go-api/internal/interfaces/http/dto"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

type mockGetUserByExternalIDHandler struct {
	calls  int
	users  []*domainuser.User
	errs   []error
}

func (m *mockGetUserByExternalIDHandler) Handle(_ context.Context, externalID string) (*domainuser.User, error) {
	idx := m.calls
	m.calls++
	if idx >= len(m.errs) {
		return nil, errors.New("unexpected get user call")
	}
	if m.errs[idx] != nil {
		return nil, m.errs[idx]
	}
	if externalID != "clerk_123" {
		return nil, errors.New("unexpected external id")
	}
	return m.users[idx], nil
}

type mockCreateUserHandler struct {
	called bool
	cmd    usercmd.CreateUserCommand
	user   *domainuser.User
	err    error
}

func (m *mockCreateUserHandler) Handle(_ context.Context, cmd usercmd.CreateUserCommand) (*domainuser.User, error) {
	m.called = true
	m.cmd = cmd
	return m.user, m.err
}

type mockUpdateUserHandler struct {
	called bool
	cmd    usercmd.UpdateUserCommand
	err    error
}

func (m *mockUpdateUserHandler) Handle(_ context.Context, cmd usercmd.UpdateUserCommand) error {
	m.called = true
	m.cmd = cmd
	return m.err
}

type mockDeleteUserByExternalIDHandler struct {
	called     bool
	externalID string
	err        error
}

func (m *mockDeleteUserByExternalIDHandler) Handle(_ context.Context, externalID string) error {
	m.called = true
	m.externalID = externalID
	return m.err
}

type userWebhookMocks struct {
	getByExternalID *mockGetUserByExternalIDHandler
	createUser      *mockCreateUserHandler
	updateUser      *mockUpdateUserHandler
	deleteUser      *mockDeleteUserByExternalIDHandler
}

func newUserWebhookHandler(mocks userWebhookMocks) *handler.UserWebhookHandler {
	if mocks.getByExternalID == nil {
		mocks.getByExternalID = &mockGetUserByExternalIDHandler{}
	}
	if mocks.createUser == nil {
		mocks.createUser = &mockCreateUserHandler{}
	}
	if mocks.updateUser == nil {
		mocks.updateUser = &mockUpdateUserHandler{}
	}
	if mocks.deleteUser == nil {
		mocks.deleteUser = &mockDeleteUserByExternalIDHandler{}
	}

	return handler.NewUserWebhookHandler(
		mocks.getByExternalID,
		mocks.createUser,
		mocks.updateUser,
		mocks.deleteUser,
	)
}

func executeWebhook(t *testing.T, h *handler.UserWebhookHandler, event dto.ClerkEvent) *http.Response {
	t.Helper()

	app := testutil.NewTestApp()
	app.Post("/webhooks/clerk", testutil.WithLocal("payload", event), h.Execute)

	req, err := testutil.JSONRequest(http.MethodPost, "/webhooks/clerk", nil)
	if err != nil {
		t.Fatalf("build request: %v", err)
	}

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}
	return resp
}

func clerkEvent(eventType string, data any) dto.ClerkEvent {
	raw, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return dto.ClerkEvent{
		Type:       eventType,
		InstanceID: "ins_123",
		Object:     "event",
		Timestamp:  1704067200,
		Data:       raw,
	}
}

func sampleUserCreatedData() map[string]any {
	return map[string]any{
		"id":         "clerk_123",
		"first_name": "Jane",
		"last_name":  "Doe",
		"banned":     false,
		"email_addresses": []map[string]string{
			{"id": "email_1", "email_address": "jane@example.com"},
		},
	}
}

func sampleUserUpdatedData() map[string]any {
	return map[string]any{
		"id":         "clerk_123",
		"first_name": "Jane",
		"last_name":  "Smith",
		"banned":     false,
		"email_addresses": []map[string]string{
			{"id": "email_1", "email_address": "jane.smith@example.com"},
		},
	}
}

func sampleDomainUser() *domainuser.User {
	return &domainuser.User{
		ID:        testutil.TestUserID,
		ClerkID:   "clerk_123",
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}
}

func TestUserWebhookHandler_Execute_UserCreated_Success(t *testing.T) {
	getByExternalID := &mockGetUserByExternalIDHandler{
		users: []*domainuser.User{nil},
		errs:  []error{nil},
	}
	createUser := &mockCreateUserHandler{
		user: &domainuser.User{ID: uuid.New()},
	}
	h := newUserWebhookHandler(userWebhookMocks{
		getByExternalID: getByExternalID,
		createUser:      createUser,
	})

	resp := executeWebhook(t, h, clerkEvent("user.created", sampleUserCreatedData()))
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusCreated)
	}
	if getByExternalID.calls != 1 {
		t.Fatalf("get user calls: got %d want 1", getByExternalID.calls)
	}
	if !createUser.called {
		t.Fatal("expected create user handler to be called")
	}
	if createUser.cmd.ClerkID != "clerk_123" {
		t.Fatalf("clerk id: got %q", createUser.cmd.ClerkID)
	}
	if createUser.cmd.Email != "jane@example.com" {
		t.Fatalf("email: got %q", createUser.cmd.Email)
	}
}

func TestUserWebhookHandler_Execute_UserCreated_HandlerError_Internal(t *testing.T) {
	getByExternalID := &mockGetUserByExternalIDHandler{
		users: []*domainuser.User{nil},
		errs:  []error{nil},
	}
	createUser := &mockCreateUserHandler{err: errors.New("database unavailable")}
	h := newUserWebhookHandler(userWebhookMocks{
		getByExternalID: getByExternalID,
		createUser:      createUser,
	})

	resp := executeWebhook(t, h, clerkEvent("user.created", sampleUserCreatedData()))
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Failed to create user" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestUserWebhookHandler_Execute_UserUpdated_Success(t *testing.T) {
	getByExternalID := &mockGetUserByExternalIDHandler{
		users: []*domainuser.User{sampleDomainUser()},
		errs:  []error{nil},
	}
	updateUser := &mockUpdateUserHandler{}
	h := newUserWebhookHandler(userWebhookMocks{
		getByExternalID: getByExternalID,
		updateUser:      updateUser,
	})

	resp := executeWebhook(t, h, clerkEvent("user.updated", sampleUserUpdatedData()))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !updateUser.called {
		t.Fatal("expected update user handler to be called")
	}
	if updateUser.cmd.Email != "jane.smith@example.com" {
		t.Fatalf("email: got %q", updateUser.cmd.Email)
	}
	if updateUser.cmd.LastName != "Smith" {
		t.Fatalf("last name: got %q", updateUser.cmd.LastName)
	}
}

func TestUserWebhookHandler_Execute_UserUpdated_UserNotFound(t *testing.T) {
	getByExternalID := &mockGetUserByExternalIDHandler{
		users: []*domainuser.User{nil},
		errs:  []error{nil},
	}
	h := newUserWebhookHandler(userWebhookMocks{getByExternalID: getByExternalID})

	resp := executeWebhook(t, h, clerkEvent("user.updated", sampleUserUpdatedData()))
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNotFound)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "User not found" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestUserWebhookHandler_Execute_UserUpdated_HandlerError_Internal(t *testing.T) {
	getByExternalID := &mockGetUserByExternalIDHandler{
		users: []*domainuser.User{sampleDomainUser()},
		errs:  []error{nil},
	}
	updateUser := &mockUpdateUserHandler{err: errors.New("database unavailable")}
	h := newUserWebhookHandler(userWebhookMocks{
		getByExternalID: getByExternalID,
		updateUser:      updateUser,
	})

	resp := executeWebhook(t, h, clerkEvent("user.updated", sampleUserUpdatedData()))
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Failed to update user" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestUserWebhookHandler_Execute_UserDeleted_Success(t *testing.T) {
	deleteUser := &mockDeleteUserByExternalIDHandler{}
	h := newUserWebhookHandler(userWebhookMocks{deleteUser: deleteUser})

	resp := executeWebhook(t, h, clerkEvent("user.deleted", map[string]string{"id": "clerk_123"}))
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusNoContent)
	}
	if !deleteUser.called {
		t.Fatal("expected delete user handler to be called")
	}
	if deleteUser.externalID != "clerk_123" {
		t.Fatalf("external id: got %q", deleteUser.externalID)
	}
}

func TestUserWebhookHandler_Execute_UserDeleted_HandlerError_Internal(t *testing.T) {
	deleteUser := &mockDeleteUserByExternalIDHandler{err: errors.New("database unavailable")}
	h := newUserWebhookHandler(userWebhookMocks{deleteUser: deleteUser})

	resp := executeWebhook(t, h, clerkEvent("user.deleted", map[string]string{"id": "clerk_123"}))
	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusInternalServerError)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Failed to delete user" {
		t.Fatalf("message: got %v", body["message"])
	}
}

func TestUserWebhookHandler_Execute_InvalidPayload_BadRequest(t *testing.T) {
	h := newUserWebhookHandler(userWebhookMocks{})

	event := dto.ClerkEvent{
		Type:       "user.created",
		InstanceID: "ins_123",
		Object:     "event",
		Timestamp:  1704067200,
		Data:       json.RawMessage(`{invalid-json`),
	}

	resp := executeWebhook(t, h, event)
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status: got %d want %d", resp.StatusCode, http.StatusBadRequest)
	}

	body := testutil.DecodeJSONMap(t, resp)
	if body["message"] != "Invalid request body" {
		t.Fatalf("message: got %v", body["message"])
	}
}
