package auth

import (
	"net/http"
	"testing"

	fberrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

type mockUserStore struct {
	users map[string]*users.User
}

func (m *mockUserStore) Get(_ string, id interface{}) (*users.User, error) {
	if v, ok := id.(string); ok {
		if u, ok := m.users[v]; ok {
			return u, nil
		}
	}
	return nil, fberrors.ErrNotExist
}

func (m *mockUserStore) Gets(_ string) ([]*users.User, error) { return nil, nil }
func (m *mockUserStore) Update(_ *users.User, _ ...string) error { return nil }
func (m *mockUserStore) Save(user *users.User) error {
	m.users[user.Username] = user
	return nil
}
func (m *mockUserStore) Delete(_ interface{}) error { return nil }
func (m *mockUserStore) LastUpdate(_ uint) int64     { return 0 }

func TestProxyAuthCreateUserRestrictsDefaults(t *testing.T) {
	t.Parallel()

	store := &mockUserStore{users: make(map[string]*users.User)}
	srv := &settings.Server{Root: t.TempDir()}

	s := &settings.Settings{
		Key:        []byte("key"),
		AuthMethod: MethodProxyAuth,
		Defaults: settings.UserDefaults{
			Perm: users.Permissions{
				Admin:    true,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
			Commands: []string{"git", "ls", "cat", "id"},
		},
	}

	auth := ProxyAuth{Header: "X-Remote-User"}
	req, _ := http.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Remote-User", "newproxyuser")

	user, err := auth.Auth(req, store, s, srv)
	if err != nil {
		t.Fatalf("Auth() error: %v", err)
	}

	if user.Perm.Admin {
		t.Error("auto-provisioned proxy user should not have Admin permission")
	}
	if user.Perm.Execute {
		t.Error("auto-provisioned proxy user should not have Execute permission")
	}
	if len(user.Commands) != 0 {
		t.Errorf("auto-provisioned proxy user should have empty Commands, got %v", user.Commands)
	}
	if !user.Perm.Create {
		t.Error("auto-provisioned proxy user should retain Create permission from defaults")
	}
}
