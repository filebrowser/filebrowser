package fbhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/settings"
	fbstorage "github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

func setupTestStorage(t *testing.T) *httpTestEnv {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := storm.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	})

	storage, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := storage.Settings.Save(&settings.Settings{
		AuthMethod: "json",
		Key:        []byte("testkey"),
		Defaults: settings.UserDefaults{
			Perm: users.Permissions{Admin: false},
		},
	}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	if err := storage.Auth.Save(&auth.JSONAuth{}); err != nil {
		t.Fatalf("failed to save auth: %v", err)
	}

	user := &users.User{
		Username: "testuser",
		Password: "",
		Perm:     users.Permissions{Admin: false},
	}
	pwd, err := users.ValidateAndHashPwd("S3cur3P@ssw0rd!xyz", 0)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user.Password = pwd
	if err := storage.Users.Save(user); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	server := &settings.Server{Root: t.TempDir()}

	return &httpTestEnv{
		storage: storage,
		server:  server,
		user:    user,
	}
}

type httpTestEnv struct {
	storage *fbstorage.Storage
	server  *settings.Server
	user    *users.User
}

func createTestToken(t *testing.T, env *httpTestEnv, userID uint, expiry time.Duration) string {
	t.Helper()
	tokenStr, err := auth.GenerateToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	tok := &auth.Token{
		Token:     tokenStr,
		UserID:    userID,
		ExpiresAt: time.Now().Add(expiry),
		CreatedAt: time.Now(),
	}
	if err := env.storage.Tokens.Save(tok); err != nil {
		t.Fatalf("failed to save token: %v", err)
	}
	return tokenStr
}

func TestExtractToken(t *testing.T) {
	t.Parallel()

	t.Run("present", func(t *testing.T) {
		t.Parallel()
		r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
		r.Header.Set("X-Auth", "my-token")
		if got := extractToken(r); got != "my-token" {
			t.Errorf("extractToken() = %q, want %q", got, "my-token")
		}
	})

	t.Run("missing", func(t *testing.T) {
		t.Parallel()
		r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
		if got := extractToken(r); got != "" {
			t.Errorf("extractToken() = %q, want empty", got)
		}
	})
}

func TestWithUser_NoToken(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	handler := handle(withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		return http.StatusOK, nil
	}), "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestWithUser_InvalidToken(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	handler := handle(withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		return http.StatusOK, nil
	}), "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Set("X-Auth", "invalid-token-value")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestWithUser_ExpiredToken(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	tokenStr := createTestToken(t, env, env.user.ID, -1*time.Hour)

	handler := handle(withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		return http.StatusOK, nil
	}), "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Set("X-Auth", tokenStr)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}

	_, err := env.storage.Tokens.Get(tokenStr)
	if err == nil {
		t.Error("expired token should have been deleted from store")
	}
}

func TestWithUser_ValidToken(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	tokenStr := createTestToken(t, env, env.user.ID, 1*time.Hour)

	var capturedUser *users.User
	handler := handle(withUser(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
		capturedUser = d.user
		return http.StatusOK, nil
	}), "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Set("X-Auth", tokenStr)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if capturedUser == nil {
		t.Fatal("user should have been set in data")
	}
	if capturedUser.Username != "testuser" {
		t.Errorf("username = %q, want %q", capturedUser.Username, "testuser")
	}
}

func TestLoginHandler(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	handler := handle(loginHandler(2*time.Hour), "", env.storage, env.server)

	body := `{"username":"testuser","password":"S3cur3P@ssw0rd!xyz"}`
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	tokenStr := recorder.Body.String()
	if len(tokenStr) != 64 {
		t.Errorf("expected 64-char token, got %d chars", len(tokenStr))
	}

	// Verify token exists in store
	tok, err := env.storage.Tokens.Get(tokenStr)
	if err != nil {
		t.Fatalf("token not found in store: %v", err)
	}
	if tok.UserID != env.user.ID {
		t.Errorf("token userID = %d, want %d", tok.UserID, env.user.ID)
	}
}

func TestLoginHandler_WrongPassword(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	handler := handle(loginHandler(2*time.Hour), "", env.storage, env.server)

	body := `{"username":"testuser","password":"wrongpassword"}`
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestLogoutHandler(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	tokenStr := createTestToken(t, env, env.user.ID, 1*time.Hour)

	handler := handle(logoutHandler, "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodPost, "/", http.NoBody)
	r.Header.Set("X-Auth", tokenStr)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	// Token should be invalidated
	_, err := env.storage.Tokens.Get(tokenStr)
	if err == nil {
		t.Error("token should have been deleted after logout")
	}
}

func TestLogoutHandler_NoToken(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	handler := handle(logoutHandler, "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodPost, "/", http.NoBody)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	// Should get 401 since withUser wraps logoutHandler
	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestRenewHandler(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	oldToken := createTestToken(t, env, env.user.ID, 1*time.Hour)

	handler := handle(renewHandler(2*time.Hour), "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodPost, "/", http.NoBody)
	r.Header.Set("X-Auth", oldToken)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	newToken := recorder.Body.String()
	if len(newToken) != 64 {
		t.Errorf("expected 64-char token, got %d chars", len(newToken))
	}

	// Old token should be invalidated
	_, err := env.storage.Tokens.Get(oldToken)
	if err == nil {
		t.Error("old token should have been deleted after renewal")
	}

	// New token should be valid
	tok, err := env.storage.Tokens.Get(newToken)
	if err != nil {
		t.Fatalf("new token not found in store: %v", err)
	}
	if tok.UserID != env.user.ID {
		t.Errorf("new token userID = %d, want %d", tok.UserID, env.user.ID)
	}
}

func TestRenewHandler_InvalidToken(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	handler := handle(renewHandler(2*time.Hour), "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodPost, "/", http.NoBody)
	r.Header.Set("X-Auth", "nonexistent-token")
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestMeHandler(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	tokenStr := createTestToken(t, env, env.user.ID, 1*time.Hour)

	handler := handle(meHandler, "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	r.Header.Set("X-Auth", tokenStr)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var info userInfo
	if err := json.NewDecoder(recorder.Body).Decode(&info); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if info.Username != "testuser" {
		t.Errorf("username = %q, want %q", info.Username, "testuser")
	}
	if info.ID != env.user.ID {
		t.Errorf("id = %d, want %d", info.ID, env.user.ID)
	}
}

func TestMeHandler_NoToken(t *testing.T) {
	t.Parallel()
	env := setupTestStorage(t)

	handler := handle(meHandler, "", env.storage, env.server)

	r, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, r)

	if recorder.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
	}
}

func TestUserInfoFrom(t *testing.T) {
	t.Parallel()

	user := &users.User{
		ID:                    42,
		Username:              "alice",
		Locale:                "en",
		ViewMode:              users.ListViewMode,
		SingleClick:           true,
		RedirectAfterCopyMove: true,
		Perm:                  users.Permissions{Admin: true, Create: true},
		LockPassword:          false,
		Commands:              []string{"ls", "cat"},
		HideDotfiles:          true,
		DateFormat:            true,
		AceEditorTheme:        "monokai",
	}

	info := userInfoFrom(user)

	if info.ID != 42 {
		t.Errorf("ID = %d, want 42", info.ID)
	}
	if info.Username != "alice" {
		t.Errorf("Username = %q, want %q", info.Username, "alice")
	}
	if info.Locale != "en" {
		t.Errorf("Locale = %q, want %q", info.Locale, "en")
	}
	if !info.SingleClick {
		t.Error("SingleClick should be true")
	}
	if !info.Perm.Admin {
		t.Error("Perm.Admin should be true")
	}
	if !info.HideDotfiles {
		t.Error("HideDotfiles should be true")
	}
	if len(info.Commands) != 2 {
		t.Errorf("Commands length = %d, want 2", len(info.Commands))
	}
}
