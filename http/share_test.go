package fbhttp

import (
	"encoding/json"
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
