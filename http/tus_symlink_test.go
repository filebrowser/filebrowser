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

	"github.com/filebrowser/filebrowser/v2/files"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
	"github.com/filebrowser/filebrowser/v2/users"
)

// Reproduces the TUS write vector of GHSA-v9g6-9pp4-3w22: a scoped user must
// not be able to create or write files through a symlinked directory that
// escapes their scope.
func TestTusHandlersRejectSymlinkScopeEscape(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	outside := filepath.Join(root, "otheruser")
	for _, d := range []string{userScope, outside} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	// A directory symlink inside the user's scope pointing outside it.
	if err := os.Symlink(outside, filepath.Join(userScope, "escape_link")); err != nil {
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
	if err := st.Users.Save(&users.User{
		Username: "u",
		Password: "pw",
		Perm:     users.Permissions{Create: true, Modify: true},
	}); err != nil {
		t.Fatalf("failed to save user: %v", err)
	}
	if err := st.Settings.Save(&settings.Settings{Key: key}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}
	st.Users = &customFSUser{
		Store: st.Users,
		fs:    files.NewScopedFs(afero.NewOsFs(), userScope),
	}

	// Forge a valid auth token for user ID 1.
	claims := &authToken{
		User: userInfo{ID: 1, Username: "u", Perm: users.Permissions{Create: true, Modify: true}},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	cases := map[string]struct {
		method  string
		handler handleFunc
		headers map[string]string
	}{
		"POST create through symlinked dir": {
			method:  http.MethodPost,
			handler: tusPostHandler(newMemoryUploadCache()),
			headers: map[string]string{"Upload-Length": "20"},
		},
		"PATCH write through symlinked dir": {
			method:  http.MethodPatch,
			handler: tusPatchHandler(newMemoryUploadCache()),
			headers: map[string]string{"Content-Type": "application/offset+octet-stream", "Upload-Offset": "0"},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, "escape_link/injected.txt", http.NoBody)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Auth", signed)
			for k, v := range tc.headers {
				req.Header.Set(k, v)
			}

			recorder := httptest.NewRecorder()
			handler := handle(tc.handler, "", st, &settings.Server{})
			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusForbidden {
				t.Errorf("expected 403, got %d", recorder.Code)
			}
			if _, statErr := os.Stat(filepath.Join(outside, "injected.txt")); statErr == nil {
				t.Errorf("VULNERABLE: file was created outside the user's scope")
			}
		})
	}
}
