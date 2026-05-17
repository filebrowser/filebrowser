package fbhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/golang-jwt/jwt/v5"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/share"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestSharePostHandlerRejectsDuplicateCustomKey(t *testing.T) {
	t.Parallel()

	testEnv := newShareTestEnv(t)
	if err := testEnv.storage.Share.Save(&share.Link{
		Hash:   "team-docs",
		Path:   "/existing.txt",
		UserID: testEnv.user.ID,
	}); err != nil {
		t.Fatalf("failed to seed share: %v", err)
	}

	req := testEnv.newShareRequest(t, map[string]string{
		"hash": "team-docs",
	})

	recorder := httptest.NewRecorder()
	testEnv.handler.ServeHTTP(recorder, req)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusConflict)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "share key already exists") {
		t.Fatalf("body = %q, want duplicate message", string(body))
	}
}

func TestSharePostHandlerReusesExpiredCustomKey(t *testing.T) {
	t.Parallel()

	testEnv := newShareTestEnv(t)
	if err := testEnv.storage.Share.Save(&share.Link{
		Hash:   "team-docs",
		Path:   "/expired.txt",
		UserID: testEnv.user.ID,
		Expire: time.Now().Add(-time.Hour).Unix(),
	}); err != nil {
		t.Fatalf("failed to seed expired share: %v", err)
	}

	req := testEnv.newShareRequest(t, map[string]string{
		"hash": "team-docs",
	})
	req.URL.Path = "/api/share/new.txt"

	recorder := httptest.NewRecorder()
	testEnv.handler.ServeHTTP(recorder, req)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusOK)
	}

	var link share.Link
	if err := json.NewDecoder(result.Body).Decode(&link); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if link.Hash != "team-docs" {
		t.Fatalf("hash = %q, want %q", link.Hash, "team-docs")
	}

	if link.Path != "/new.txt" {
		t.Fatalf("path = %q, want %q", link.Path, "/new.txt")
	}

	stored, err := testEnv.storage.Share.GetByHash("team-docs")
	if err != nil {
		t.Fatalf("failed to reload share: %v", err)
	}

	if stored.Path != "/new.txt" {
		t.Fatalf("stored path = %q, want %q", stored.Path, "/new.txt")
	}
}

func TestSharePostHandlerRejectsInvalidCustomKey(t *testing.T) {
	t.Parallel()

	testEnv := newShareTestEnv(t)

	req := testEnv.newShareRequest(t, map[string]string{
		"hash": "team docs",
	})

	recorder := httptest.NewRecorder()
	testEnv.handler.ServeHTTP(recorder, req)

	result := recorder.Result()
	defer result.Body.Close()

	if result.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", result.StatusCode, http.StatusBadRequest)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	if !strings.Contains(string(body), "share key must use only letters, numbers, hyphen, or underscore") {
		t.Fatalf("body = %q, want validation message", string(body))
	}
}

type shareTestEnv struct {
	handler http.Handler
	storage *storage.Storage
	user    *users.User
	token   string
}

func newShareTestEnv(t *testing.T) *shareTestEnv {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "db")
	db, err := storm.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("failed to close db: %v", err)
		}
	})

	store, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	settingsKey := []byte("share-test-key")
	if err := store.Settings.Save(&settings.Settings{Key: settingsKey}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	user := &users.User{
		Username: "share-user",
		Password: "pw",
		Perm: users.Permissions{
			Share:    true,
			Download: true,
		},
	}
	if err := store.Users.Save(user); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}

	handler := handle(sharePostHandler, "/api/share", store, &settings.Server{})
	return &shareTestEnv{
		handler: handler,
		storage: store,
		user:    user,
		token:   newShareTestToken(t, settingsKey, user),
	}
}

func (e *shareTestEnv) newShareRequest(t *testing.T, body map[string]string) *http.Request {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "/api/share/file.txt", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Auth", e.token)
	return req
}

func newShareTestToken(t *testing.T, key []byte, user *users.User) string {
	t.Helper()

	claims := &authToken{
		User: userInfo{
			ID:       user.ID,
			Perm:     user.Perm,
			Username: user.Username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Issuer:    "File Browser",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return signed
}
