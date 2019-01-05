package auth

import (
	"net/http"

	"github.com/filebrowser/filebrowser/v2/users"
)

// Auther is the authentication interface.
type Auther interface {
	// Auth is called to authenticate a request.
	Auth(*http.Request) (*users.User, error)
	// SetStorage attaches the Storage instance.
	SetStorage(*users.Storage)
}
