package lib

import "net/http"

// AuthMethod describes an authentication method.
type AuthMethod string

// Auther is the authentication interface.
type Auther interface {
	// Auth is called to authenticate a request.
	Auth(*http.Request) (*User, error)
	// SetInstance attaches the File Browser instance.
	SetInstance(*FileBrowser)
}
