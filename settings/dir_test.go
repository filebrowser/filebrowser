package settings

import (
	"errors"
	"testing"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/users"
)

// stubStore is a minimal users.Store used to exercise CreateUserHome.
type stubStore struct {
	byScope map[string]*users.User
}

func (s *stubStore) Get(_ string, _ bool, _ interface{}) (*users.User, error) {
	return nil, fberrors.ErrNotExist
}

func (s *stubStore) GetByScope(scope string) (*users.User, error) {
	if u, ok := s.byScope[scope]; ok {
		return u, nil
	}
	return nil, fberrors.ErrNotExist
}

func (s *stubStore) Gets(_ string, _ bool) ([]*users.User, error) { return nil, nil }
func (s *stubStore) Update(_ *users.User, _ ...string) error      { return nil }
func (s *stubStore) Save(_ *users.User) error                     { return nil }
func (s *stubStore) Delete(_ interface{}) error                   { return nil }
func (s *stubStore) LastUpdate(_ uint) int64                      { return 0 }

// A user provisioned with CreateUserDir must receive a per-user home directory
// derived from its username, not the default scope which normalizes to the
// server root.
func TestCreateUserHomeDerivesPerUserScope(t *testing.T) {
	s := &Settings{CreateUserDir: true, UserHomeBasePath: "/users"}
	store := &stubStore{byScope: map[string]*users.User{}}

	user := &users.User{Username: "alice", Scope: "."}
	if err := s.CreateUserHome(user, store, t.TempDir(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Scope != "/users/alice" {
		t.Errorf("expected derived scope /users/alice, got %q", user.Scope)
	}
}

// When the derived scope is already owned by another user, provisioning must be
// rejected so distinct usernames cannot silently share one home directory.
func TestCreateUserHomeRejectsCollision(t *testing.T) {
	s := &Settings{CreateUserDir: true, UserHomeBasePath: "/users"}
	store := &stubStore{byScope: map[string]*users.User{"/users/alice": {Username: "alice"}}}

	user := &users.User{Username: "alice", Scope: "."}
	if err := s.CreateUserHome(user, store, t.TempDir(), false); !errors.Is(err, fberrors.ErrExist) {
		t.Fatalf("expected ErrExist on scope collision, got %v", err)
	}
}

// A scope explicitly supplied by the caller (e.g. returned by an auth hook) must
// be preserved instead of being replaced by a derived home directory.
func TestCreateUserHomePreservesExplicitScope(t *testing.T) {
	s := &Settings{CreateUserDir: true, UserHomeBasePath: "/users"}
	store := &stubStore{byScope: map[string]*users.User{"/users/alice": {Username: "alice"}}}

	user := &users.User{Username: "alice", Scope: "/custom"}
	if err := s.CreateUserHome(user, store, t.TempDir(), true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Scope != "/custom" {
		t.Errorf("explicit scope should be preserved, got %q", user.Scope)
	}
}
