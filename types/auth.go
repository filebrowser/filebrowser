package types

import "net/http"

// Auther is the authentication interface.
type Auther interface {
	// Auth is called to authenticate a request.
	Auth(*http.Request) (*User, error)
	// SetStorage gives the Auther the storage.
	SetStorage(*Storage)
}
