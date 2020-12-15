package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/filebrowser/filebrowser/v2/settings"
)

// Auther is the authentication interface.
type Auther interface {
	// Auth is called to authenticate a request.
	Auth(r *http.Request, s *users.Storage, set *settings.Storage, root string) (*users.User, error)
	// LoginPage indicates if this auther needs a login page.
	LoginPage() bool
}
