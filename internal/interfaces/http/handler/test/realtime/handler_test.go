package realtimetest

import (
	"errors"
	"net/http"
	"testing"

	"go-api/internal/infrastructure/centrifugo"
	"go-api/internal/interfaces/http/handler"
	"go-api/internal/interfaces/http/testutil"

	"github.com/google/uuid"
)

type mockConnectionCreator struct {
	called bool
	userID uuid.UUID
	info   centrifugo.ConnectionInfo
	err    error
}

func (m *mockConnectionCreator) CreateConnectionInfo(userID uuid.UUID) (centrifugo.ConnectionInfo, error) {
	m.called = true
	m.userID = userID
	return m.info, m.err
}

func newRealtimeHandler(creator *mockConnectionCreator) *handler.RealtimeHandler {
	if creator == nil {
		creator = &mockConnectionCreator{}
	}
	return handler.NewRealtimeHandler(creator)
}

func TestRealtimeHandler_GetConnection_Success(t *testing.T) {
	creator := &mockConnectionCreator{
		info: centrifugo.ConnectionInfo{
			Token:   "signed-token",
			Channel: centrifugo.UserChannel(testutil.TestUserID),
			WSURL:   "wss://realtime.example.com/connection/websocket",
		},
	}
	h := newRealtimeHandler(creator)

	app := testutil.NewTestApp()
	app.Get("/realtime/connection", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetConnection)

	req, err := testutil.JSONRequest(http.MethodGet, "/realtime/connection", nil)
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
	if !creator.called {
		t.Fatal("expected connection creator to be called")
	}
	if creator.userID != testutil.TestUserID {
		t.Fatalf("user id: got %s", creator.userID)
	}

	var out centrifugo.ConnectionInfo
	testutil.DecodeJSON(t, resp, &out)
	if out.Token != "signed-token" {
		t.Fatalf("token: got %q", out.Token)
	}
	if out.Channel != centrifugo.UserChannel(testutil.TestUserID) {
		t.Fatalf("channel: got %q", out.Channel)
	}
}

func TestRealtimeHandler_GetConnection_Unauthorized(t *testing.T) {
	creator := &mockConnectionCreator{}
	h := newRealtimeHandler(creator)

	app := testutil.NewTestApp()
	app.Get("/realtime/connection", h.GetConnection)

	req, err := testutil.JSONRequest(http.MethodGet, "/realtime/connection", nil)
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
	if creator.called {
		t.Fatal("connection creator must not be called without auth")
	}
}

func TestRealtimeHandler_GetConnection_HandlerError_Internal(t *testing.T) {
	creator := &mockConnectionCreator{err: errors.New("failed to sign token")}
	h := newRealtimeHandler(creator)

	app := testutil.NewTestApp()
	app.Get("/realtime/connection", testutil.WithActiveProject(testutil.TestUserID, testutil.TestProjectID), h.GetConnection)

	req, err := testutil.JSONRequest(http.MethodGet, "/realtime/connection", nil)
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
	if !creator.called {
		t.Fatal("expected connection creator to be called")
	}
}
