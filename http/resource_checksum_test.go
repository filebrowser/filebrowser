package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// The ?checksum= branch reads the whole file to compute a digest, so it must be
// gated behind Perm.Download like the other read paths. A user without download
// permission must not obtain a digest of a file they cannot download.
func TestResourceChecksumRequiresDownloadPermission(t *testing.T) {
	root := t.TempDir()
	userScope := filepath.Join(root, "user")
	if err := os.MkdirAll(userScope, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(userScope, "secret.txt"), []byte("classified"), 0o644); err != nil {
		t.Fatal(err)
	}
	key := []byte("test-signing-key")

	get := func(t *testing.T, perm users.Permissions) *httptest.ResponseRecorder {
		st := scopedUserStorage(t, userScope, perm, key)
		req, _ := http.NewRequest(http.MethodGet, "/secret.txt?checksum=sha256", http.NoBody)
		req.Header.Set("X-Auth", signToken(t, perm, key))
		rec := httptest.NewRecorder()
		handle(resourceGetHandler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		return rec
	}

	t.Run("denied without download permission", func(t *testing.T) {
		rec := get(t, users.Permissions{})
		if rec.Code != http.StatusAccepted {
			t.Fatalf("expected 202, got %d body=%q", rec.Code, rec.Body.String())
		}
		if strings.Contains(strings.ToLower(rec.Body.String()), "checksum") {
			t.Fatalf("digest leaked without download permission: %q", rec.Body.String())
		}
	})

	t.Run("allowed with download permission", func(t *testing.T) {
		rec := get(t, users.Permissions{Download: true})
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d body=%q", rec.Code, rec.Body.String())
		}
		if !strings.Contains(rec.Body.String(), "checksums") {
			t.Fatalf("expected a checksum in the response, got %q", rec.Body.String())
		}
	})
}
