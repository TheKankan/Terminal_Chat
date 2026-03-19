package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TheKankan/TerminalSecuredChat/internal/database"
	"github.com/google/uuid"
)

// ── Mock ──────────────────────────────────────────────────────────────────────

type mockStore struct {
	users map[string]database.User
}

func newMockStore() *mockStore {
	return &mockStore{users: make(map[string]database.User)}
}

func (m *mockStore) CreateUser(_ context.Context, arg database.CreateUserParams) (database.User, error) {
	user := database.User{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Username:       arg.Username,
		HashedPassword: arg.HashedPassword,
	}
	m.users[arg.Username] = user
	return user, nil
}

func (m *mockStore) GetUserFromUsername(_ context.Context, username string) (database.User, error) {
	user, ok := m.users[username]
	if !ok {
		return database.User{}, &userNotFoundError{username: username}
	}
	return user, nil
}

func (m *mockStore) GetUsernameFromID(_ context.Context, id uuid.UUID) (string, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u.Username, nil
		}
	}
	return "", &userNotFoundError{}
}

func (m *mockStore) ChangePasswordAndUsername(_ context.Context, arg database.ChangePasswordAndUsernameParams) (database.User, error) {
	for username, u := range m.users {
		if u.ID == arg.ID {
			delete(m.users, username)
			u.Username = arg.Username
			u.HashedPassword = arg.HashedPassword
			m.users[arg.Username] = u
			return u, nil
		}
	}
	return database.User{}, &userNotFoundError{}
}

func (m *mockStore) CreateMessage(_ context.Context, arg database.CreateMessageParams) (database.Message, error) {
	return database.Message{}, nil
}

func (m *mockStore) GetRecentMessages(_ context.Context, limit int32) ([]database.GetRecentMessagesRow, error) {
	return []database.GetRecentMessagesRow{}, nil
}

type userNotFoundError struct{ username string }

func (e *userNotFoundError) Error() string { return "user not found: " + e.username }

// ── Helpers ───────────────────────────────────────────────────────────────────

func newTestConfig(store *mockStore) *apiConfig {
	return &apiConfig{
		db:        store,
		jwtSecret: "test-secret",
	}
}

func makeRequest(t *testing.T, method, path string, body interface{}) *http.Request {
	t.Helper()
	jsonBody, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	return req
}

// ── handlerRegister ───────────────────────────────────────────────────────────

func TestHandlerRegister_Success(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/register", map[string]string{
		"username": "alice",
		"password": "secret123",
	})

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		User  User   `json:"user"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Username != "alice" {
		t.Errorf("expected username 'alice', got '%s'", resp.User.Username)
	}
	if resp.Token == "" {
		t.Error("expected a non-empty token")
	}
}

func TestHandlerRegister_MissingUsername(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/register", map[string]string{
		"username": "",
		"password": "secret123",
	})

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerRegister_MissingPassword(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/register", map[string]string{
		"username": "alice",
		"password": "",
	})

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerRegister_InvalidJSON(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")

	cfg.handlerRegister(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

// ── handlerLogin ──────────────────────────────────────────────────────────────

func TestHandlerLogin_Success(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	// Register a user first
	registerReq := makeRequest(t, http.MethodPost, "/register", map[string]string{
		"username": "alice",
		"password": "secret123",
	})
	cfg.handlerRegister(httptest.NewRecorder(), registerReq)

	// Now login
	w := httptest.NewRecorder()
	loginReq := makeRequest(t, http.MethodPost, "/login", map[string]string{
		"username": "alice",
		"password": "secret123",
	})

	cfg.handlerLogin(w, loginReq)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		User  User   `json:"user"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Username != "alice" {
		t.Errorf("expected username 'alice', got '%s'", resp.User.Username)
	}
	if resp.Token == "" {
		t.Error("expected a non-empty token")
	}
}

func TestHandlerLogin_WrongPassword(t *testing.T) {
	store := newMockStore()
	cfg := newTestConfig(store)

	// Register a user first
	registerReq := makeRequest(t, http.MethodPost, "/register", map[string]string{
		"username": "alice",
		"password": "secret123",
	})
	cfg.handlerRegister(httptest.NewRecorder(), registerReq)

	// Login with wrong password
	w := httptest.NewRecorder()
	loginReq := makeRequest(t, http.MethodPost, "/login", map[string]string{
		"username": "alice",
		"password": "wrongpassword",
	})

	cfg.handlerLogin(w, loginReq)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerLogin_UserNotFound(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/login", map[string]string{
		"username": "nobody",
		"password": "secret123",
	})

	cfg.handlerLogin(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestHandlerLogin_MissingCredentials(t *testing.T) {
	cfg := newTestConfig(newMockStore())
	w := httptest.NewRecorder()

	req := makeRequest(t, http.MethodPost, "/login", map[string]string{
		"username": "",
		"password": "",
	})

	cfg.handlerLogin(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
