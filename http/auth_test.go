package fbhttp

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/asdine/storm/v3"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage/bolt"
)

// Regression for the username-normalization home-directory collision
// (GHSA-7rc3-g7h6-22m7): with Signup and CreateUserDir enabled, two distinct
// usernames that cleanUsername() normalizes to the same directory must not be
// handed the same home directory. The second registration is rejected.
func TestSignupRejectsCollidingNormalizedScope(t *testing.T) {
	root := t.TempDir()

	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := bolt.NewStorage(db)
	if err != nil {
		t.Fatalf("failed to get storage: %v", err)
	}
	if err := st.Settings.Save(&settings.Settings{
		Key:                   []byte("test-signing-key"),
		Signup:                true,
		CreateUserDir:         true,
		UserHomeBasePath:      "/users",
		MinimumPasswordLength: 1,
	}); err != nil {
		t.Fatalf("failed to save settings: %v", err)
	}

	server := &settings.Server{Root: root}

	signup := func(username string) *httptest.ResponseRecorder {
		body := `{"username":"` + username + `","password":"CollidePw12345!"}`
		req, _ := http.NewRequest(http.MethodPost, "/signup", strings.NewReader(body))
		rec := httptest.NewRecorder()
		handle(signupHandler, "", st, server).ServeHTTP(rec, req)
		return rec
	}

	// Victim registers first and gets /users/teamone-x.
	if rec := signup("teamone-x"); rec.Code != http.StatusOK {
		t.Fatalf("first signup: expected 200, got %d body=%q", rec.Code, rec.Body.String())
	}

	// Attacker picks a distinct username that normalizes to the same scope.
	if rec := signup("teamone/x"); rec.Code != http.StatusConflict {
		t.Fatalf("VULNERABLE: colliding signup expected 409, got %d body=%q", rec.Code, rec.Body.String())
	}

	// The shared scope must still be owned solely by the first user.
	owner, err := st.Users.GetByScope("/users/teamone-x")
	if err != nil {
		t.Fatalf("expected first user to own the scope: %v", err)
	}
	if owner.Username != "teamone-x" {
		t.Fatalf("scope owner = %q, want teamone-x", owner.Username)
	}
}
