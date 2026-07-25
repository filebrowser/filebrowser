package bolt

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/asdine/storm/v3"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

// GetByScope must match case-insensitively: on a case-insensitive filesystem
// two scopes that differ only in case resolve to the same home directory, so
// the provisioning collision check has to treat them as the same scope.
func TestGetByScopeCaseInsensitive(t *testing.T) {
	db, err := storm.Open(filepath.Join(t.TempDir(), "db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	st, err := NewStorage(db)
	if err != nil {
		t.Fatalf("new storage: %v", err)
	}
	if err := st.Users.Save(&users.User{Username: "CaseVictim", Password: "pw", Scope: "/users/CaseVictim"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	for _, scope := range []string{"/users/CaseVictim", "/users/casevictim", "/USERS/CASEVICTIM"} {
		u, err := st.Users.GetByScope(scope)
		if err != nil {
			t.Errorf("GetByScope(%q) unexpected error: %v", scope, err)
			continue
		}
		if u.Username != "CaseVictim" {
			t.Errorf("GetByScope(%q) returned %q, want CaseVictim", scope, u.Username)
		}
	}

	// A genuinely different scope must still miss (and the regex-escaped query
	// must not match a superstring).
	for _, scope := range []string{"/users/other", "/users/CaseVictimX"} {
		if _, err := st.Users.GetByScope(scope); !errors.Is(err, fberrors.ErrNotExist) {
			t.Errorf("GetByScope(%q) expected ErrNotExist, got %v", scope, err)
		}
	}
}
