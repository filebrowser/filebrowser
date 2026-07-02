package fbhttp

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
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

func TestAdminShareGetsHandlerMatchesOwnerScope(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	ownerScope := filepath.Join(root, "owner")
	if err := os.MkdirAll(ownerScope, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ownerScope, "file.txt"), []byte("shared"), 0o600); err != nil {
		t.Fatal(err)
	}

	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}

	owner := &users.User{
		Username: "owner",
		Password: "pw",
		Scope:    "/owner",
		Perm:     users.Permissions{Share: true, Download: true},
	}
	if err := st.Users.Save(owner); err != nil {
		t.Fatalf("failed to save owner: %v", err)
	}

	adminPerm := users.Permissions{Admin: true, Share: true, Download: true}
	admin := &users.User{
		Username: "admin",
		Password: "pw",
		Scope:    "/",
		Perm:     adminPerm,
	}
	if err := st.Users.Save(admin); err != nil {
		t.Fatalf("failed to save admin: %v", err)
	}

	if err := st.Share.Save(&share.Link{Hash: "h", UserID: owner.ID, Path: "/file.txt"}); err != nil {
		t.Fatalf("failed to save share: %v", err)
	}
	key := []byte("test-signing-key")
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/owner/file.txt", http.NoBody)
	if err != nil {
		t.Fatalf("failed to construct request: %v", err)
	}
	req.Header.Set("X-Auth", signShareTestToken(t, admin.ID, admin.Username, adminPerm, key))

	rec := httptest.NewRecorder()
	handle(shareGetsHandler, "", st, &settings.Server{Root: root}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var links []*share.Link
	if err := json.Unmarshal(rec.Body.Bytes(), &links); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(links) != 1 || links[0].Hash != "h" {
		t.Fatalf("expected admin to see owner share h, got %#v", links)
	}
}

// Regression for the share secret exposure (GHSA-833g-cqhp-h72j): the share API
// must not serialize the bcrypt password hash or the bypass token, while still
// persisting them server-side so password-protected shares keep working.
func TestSharePostHandlerDoesNotLeakSecrets(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	if err := os.MkdirAll(userScope, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userScope, "file.txt"), []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Share: true, Download: true}
	st := scopedUserStorage(t, userScope, perm, key)
	signed := signToken(t, perm, key)

	body := `{"password":"ShareSecret123!","expires":"24","unit":"hours"}`
	req, _ := http.NewRequest(http.MethodPost, "/file.txt", strings.NewReader(body))
	req.Header.Set("X-Auth", signed)
	rec := httptest.NewRecorder()
	handle(sharePostHandler, "", st, &settings.Server{Root: root}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	var resp map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := resp["password_hash"]; ok {
		t.Errorf("VULNERABLE: response leaks password_hash: %s", rec.Body.String())
	}
	if _, ok := resp["token"]; ok {
		t.Errorf("VULNERABLE: response leaks token: %s", rec.Body.String())
	}
	if resp["hasPassword"] != true {
		t.Errorf("expected hasPassword=true, got %v", resp["hasPassword"])
	}

	// The secrets must still be persisted server-side (storm uses the JSON codec,
	// so the storage struct's tags must keep emitting them).
	stored, err := st.Share.GetByHash(resp["hash"].(string))
	if err != nil {
		t.Fatalf("share not stored: %v", err)
	}
	if stored.PasswordHash == "" || stored.Token == "" {
		t.Fatalf("server-side secrets not persisted: hash=%q token=%q", stored.PasswordHash, stored.Token)
	}
}

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

	var link shareResponse
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

	root := t.TempDir()
	for _, name := range []string{"file.txt", "new.txt"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte("shared"), 0o600); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

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

	handler := handle(sharePostHandler, "/api/share", store, &settings.Server{Root: root})
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

func signShareTestToken(t *testing.T, id uint, username string, perm users.Permissions, key []byte) string {
	t.Helper()

	claims := &authToken{
		User: userInfo{ID: id, Username: username, Perm: perm},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return signed
}
