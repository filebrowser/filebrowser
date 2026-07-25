package settings

import (
	"testing"

	"github.com/filebrowser/filebrowser/v2/users"
)

// A user provisioned with CreateUserDir must receive a per-user home directory
// derived from its username, not the default scope which normalizes to the
// server root.
func TestCreateUserHomeDerivesPerUserScope(t *testing.T) {
	s := &Settings{CreateUserDir: true, UserHomeBasePath: "/users"}

	user := &users.User{Username: "alice", Scope: "."}
	derived, err := s.CreateUserHome(user, t.TempDir(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !derived {
		t.Error("expected the scope to be reported as derived")
	}
	if user.Scope != "/users/alice" {
		t.Errorf("expected derived scope /users/alice, got %q", user.Scope)
	}
}

// A scope explicitly supplied by the caller (e.g. returned by an auth hook) must
// be preserved instead of being replaced by a derived home directory, and must
// not be reported as derived: it is legitimate for several users to share it.
func TestCreateUserHomePreservesExplicitScope(t *testing.T) {
	s := &Settings{CreateUserDir: true, UserHomeBasePath: "/users"}

	user := &users.User{Username: "alice", Scope: "/custom"}
	derived, err := s.CreateUserHome(user, t.TempDir(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if derived {
		t.Error("an explicit scope must not be reported as derived")
	}
	if user.Scope != "/custom" {
		t.Errorf("explicit scope should be preserved, got %q", user.Scope)
	}
}
