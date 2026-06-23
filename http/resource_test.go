package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/asdine/storm/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
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

// scopedUserStorage returns a storage whose single user (ID 1) is scoped to
// userScope through a symlink-confining ScopedFs (via customFSUser), mirroring
// production. Used by the symlink scope-escape regression tests below.
func scopedUserStorage(t *testing.T, userScope string, perm users.Permissions, key []byte) *storage.Storage {
	t.Helper()
	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}
	if err := st.Users.Save(&users.User{Username: "u", Password: "pw", Perm: perm}); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	st.Users = &customFSUser{
		Store: st.Users,
		fs:    afero.NewBasePathFs(afero.NewOsFs(), userScope),
	}
	return st
}

// Regression for the dangling-symlink write escape (GHSA-8wc8-hf36-mjh9 /
// GHSA-fh54-6rfh-r8f3): POSTing to an in-scope dangling symlink whose target is
// outside the scope must not dereference the link to create the out-of-scope
// file.
func TestResourcePostRejectsDanglingSymlinkWriteEscape(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	outside := filepath.Join(root, "outside")
	for _, d := range []string{userScope, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A dangling symlink inside the scope pointing at a not-yet-existing file
	// outside it (planted out-of-band, per the advisory preconditions).
	outsideTarget := filepath.Join(outside, "created.txt")
	if err := os.Symlink(outsideTarget, filepath.Join(userScope, "evil")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	key := []byte("test-signing-key")
	perm := users.Permissions{Create: true, Modify: true}
	st := scopedUserStorage(t, userScope, perm, key)
	signed := signToken(t, perm, key)

	req, _ := http.NewRequest(http.MethodPost, "/evil?override=true", strings.NewReader("http-outside"))
	req.Header.Set("X-Auth", signed)
	rec := httptest.NewRecorder()
	handle(resourcePostHandler(diskcache.NewNoOp()), "", st, &settings.Server{}).ServeHTTP(rec, req)

	if _, statErr := os.Stat(outsideTarget); statErr == nil {
		data, _ := os.ReadFile(outsideTarget)
		t.Fatalf("VULNERABLE: out-of-scope file created via dangling symlink (status=%d, content=%q)", rec.Code, string(data))
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d body=%q", rec.Code, rec.Body.String())
	}
}

// Regression for the symlink-following delete escape (GHSA-hq4g-mpch-f9vp /
// GHSA-fmm7-x4gx-8jhr): a Create-only user POSTing to a child of an escaping
// symlinked directory must not delete the out-of-scope target through the
// failed-upload cleanup RemoveAll.
func TestResourcePostCleanupDoesNotDeleteThroughSymlink(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	outside := filepath.Join(root, "outside")
	for _, d := range []string{userScope, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	victim := filepath.Join(outside, "victim.txt")
	if err := os.WriteFile(victim, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	// An escaping directory symlink inside the scope (planted out-of-band).
	if err := os.Symlink(outside, filepath.Join(userScope, "link")); err != nil {
		t.Skipf("cannot create symlink: %v", err)
	}

	key := []byte("test-signing-key")
	// Create-only: Perm.Delete is deliberately false — the bug must not need it.
	perm := users.Permissions{Create: true}
	st := scopedUserStorage(t, userScope, perm, key)
	signed := signToken(t, perm, key)

	req, _ := http.NewRequest(http.MethodPost, "/link/victim.txt", strings.NewReader("x"))
	req.Header.Set("X-Auth", signed)
	rec := httptest.NewRecorder()
	handle(resourcePostHandler(diskcache.NewNoOp()), "", st, &settings.Server{}).ServeHTTP(rec, req)

	if _, statErr := os.Stat(victim); statErr != nil {
		t.Fatalf("VULNERABLE: out-of-scope victim.txt deleted by cleanup RemoveAll (status=%d): %v", rec.Code, statErr)
	}
}
