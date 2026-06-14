package fbhttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
