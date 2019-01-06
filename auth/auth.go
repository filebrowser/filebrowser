package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

// Auther is the authentication interface.
type Auther interface {
	// Auth is called to authenticate a request.
	Auth(*http.Request, *users.Storage, *settings.Settings) (*users.User, error)
}
