package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/filebrowser/filebrowser/v2/diskcache"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

// Regression for GHSA-77x8-73f4-5485: copy, rename and delete authorize only
// the root of the operation and then recurse through the whole subtree, so a
// rule denying /src/secret used to be bypassed by copying, moving or deleting
// the allowed parent /src.
func TestRecursiveOperationsEnforceDescendantRules(t *testing.T) {
	key := []byte("test-signing-key")
	perm := users.Permissions{Create: true, Modify: true, Rename: true, Delete: true, Download: true}

	// scope builds a user scope holding an allowed parent with a denied child,
	// plus an entirely allowed tree to prove legitimate operations still run.
	scope := func(t *testing.T) (userScope string, st *storage.Storage) {
		t.Helper()
		userScope = t.TempDir()
		for _, dir := range []string{filepath.Join("src", "secret"), "clean"} {
			if err := os.MkdirAll(filepath.Join(userScope, dir), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		files := map[string]string{
			filepath.Join("src", "public.txt"):           "public",
			filepath.Join("src", "secret", "marker.txt"): "SECRET",
			filepath.Join("clean", "file.txt"):           "clean",
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(userScope, name), []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return userScope, denyRuleStorage(t, userScope, "/src/secret", perm, key)
	}

	do := func(t *testing.T, st *storage.Storage, handler handleFunc, method, target string) *httptest.ResponseRecorder {
		t.Helper()
		req, _ := http.NewRequest(method, target, http.NoBody)
		req.Header.Set("X-Auth", signToken(t, perm, key))
		rec := httptest.NewRecorder()
		handle(handler, "", st, &settings.Server{}).ServeHTTP(rec, req)
		return rec
	}

	exists := func(t *testing.T, path string) bool {
		t.Helper()
		_, err := os.Stat(path)
		return err == nil
	}

	t.Run("denied descendant is not directly readable", func(t *testing.T) {
		_, st := scope(t)
		if rec := do(t, st, rawHandler, http.MethodGet, "/src/secret/marker.txt"); rec.Code != http.StatusForbidden {
			t.Fatalf("GET /src/secret/marker.txt = %d; want 403", rec.Code)
		}
	})

	t.Run("copying the parent does not expose it", func(t *testing.T) {
		userScope, st := scope(t)
		rec := do(t, st, resourcePatchHandler(diskcache.NewNoOp()), http.MethodPatch, "/src?action=copy&destination=/dst")

		if leaked := filepath.Join(userScope, "dst", "secret", "marker.txt"); exists(t, leaked) {
			t.Fatalf("VULNERABLE: denied descendant copied to %s (status %d)", leaked, rec.Code)
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("PATCH /src?action=copy = %d; want 403", rec.Code)
		}
	})

	t.Run("renaming the parent does not expose it", func(t *testing.T) {
		userScope, st := scope(t)
		rec := do(t, st, resourcePatchHandler(diskcache.NewNoOp()), http.MethodPatch, "/src?action=rename&destination=/moved")

		if leaked := filepath.Join(userScope, "moved", "secret", "marker.txt"); exists(t, leaked) {
			t.Fatalf("VULNERABLE: denied descendant moved to %s (status %d)", leaked, rec.Code)
		}
		if !exists(t, filepath.Join(userScope, "src", "secret", "marker.txt")) {
			t.Error("denied descendant no longer at its original path")
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("PATCH /src?action=rename = %d; want 403", rec.Code)
		}
	})

	t.Run("deleting the parent does not delete it", func(t *testing.T) {
		userScope, st := scope(t)
		rec := do(t, st, resourceDeleteHandler(diskcache.NewNoOp()), http.MethodDelete, "/src")

		if !exists(t, filepath.Join(userScope, "src", "secret", "marker.txt")) {
			t.Fatalf("VULNERABLE: denied descendant deleted through its parent (status %d)", rec.Code)
		}
		if rec.Code != http.StatusForbidden {
			t.Errorf("DELETE /src = %d; want 403", rec.Code)
		}
	})

	t.Run("an entirely allowed tree still copies", func(t *testing.T) {
		userScope, st := scope(t)
		rec := do(t, st, resourcePatchHandler(diskcache.NewNoOp()), http.MethodPatch, "/clean?action=copy&destination=/clean-copy")

		if rec.Code != http.StatusOK {
			t.Fatalf("PATCH /clean?action=copy = %d, body=%q; want 200", rec.Code, rec.Body.String())
		}
		if !exists(t, filepath.Join(userScope, "clean-copy", "file.txt")) {
			t.Error("allowed tree was not copied")
		}
	})
}
