package types

import "net/http"

// Auther is the interface each authentication method must
// implement.
type Auther interface {
	Auth(r *http.Request) (*User, error)
}
