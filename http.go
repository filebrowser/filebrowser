package filemanager

import "net/http"

// StaticGen is a static website generator.
type StaticGen interface {
	SettingsPath() string
	Name() string
	Setup() error

	Hook(c *Context, w http.ResponseWriter, r *http.Request) (int, error)
	Preview(c *Context, w http.ResponseWriter, r *http.Request) (int, error)
	Publish(c *Context, w http.ResponseWriter, r *http.Request) (int, error)
}

// Context contains the needed information to make handlers work.
type Context struct {
	*FileManager
	User *User
	File *File
	// On API handlers, Router is the APi handler we want.
	Router string
}
