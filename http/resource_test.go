package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

func TestResourceCopyDoesNotDereferenceEscapingSymlink(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	if err := os.MkdirAll(filepath.Join(userScope, "srcdir"), 0o755); err != nil {
		t.Fatal(err)
	}

	// An ordinary in-scope file, to prove a normal copy still works.
	if err := os.WriteFile(filepath.Join(userScope, "srcdir", "normal.txt"), []byte("in-scope"), 0o644); err != nil {
		t.Fatal(err)
	}

	// The secret living outside the user's scope.
	secret := filepath.Join(root, "secret.txt")
	if err := os.WriteFile(secret, []byte("OUT-OF-SCOPE-SECRET"), 0o644); err != nil {
		t.Fatal(err)
	}

	// An escaping symlink planted inside the user's scope (out-of-band, as the
	// advisory's preconditions describe).
	if err := os.Symlink(secret, filepath.Join(userScope, "srcdir", "link.txt")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	key := []byte("test-signing-key")

	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}
	perm := users.Permissions{Create: true, Modify: true, Download: true, Rename: true}
	if err := st.Users.Save(&users.User{Username: "u", Password: "pw", Perm: perm}); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	// The user's scope is the real on-disk userScope. customFSUser.Get wraps it
	// in a ScopedFs, mirroring production (users.User init).
	st.Users = &customFSUser{
		Store: st.Users,
		fs:    afero.NewBasePathFs(afero.NewOsFs(), userScope),
	}

	signed := signToken(t, perm, key)

	t.Run("direct raw read is forbidden", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/srcdir/link.txt", http.NoBody)
		req.Header.Set("X-Auth", signed)
		rec := httptest.NewRecorder()
		handle(rawHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		if rec.Code == http.StatusOK {
			t.Fatalf("VULNERABLE: direct raw read of escaping symlink returned 200, body=%q", rec.Body.String())
		}
	})

	t.Run("recursive copy does not exfiltrate", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, "/srcdir/", http.NoBody)
		q := req.URL.Query()
		q.Set("action", "copy")
		q.Set("destination", "/dstdir")
		req.URL.RawQuery = q.Encode()
		req.Header.Set("X-Auth", signed)

		rec := httptest.NewRecorder()
		handle(resourcePatchHandler(diskcache.NewNoOp()), "", st, &settings.Server{}).ServeHTTP(rec, req)
		t.Logf("copy status=%d body=%q", rec.Code, rec.Body.String())

		// The escaping symlink's target content must never appear in scope.
		leaked := filepath.Join(userScope, "dstdir", "link.txt")
		if data, readErr := os.ReadFile(leaked); readErr == nil {
			t.Fatalf("VULNERABLE: out-of-scope content landed in scope at %s: %q", leaked, string(data))
		}
	})
}

func signToken(t *testing.T, perm users.Permissions, key []byte) string {
	t.Helper()
	claims := &authToken{
		User: userInfo{ID: 1, Username: "u", Perm: perm},
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
