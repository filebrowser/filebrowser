package audit

import (
	"github.com/filebrowser/filebrowser/v2/users"
)

type ResourceActivity struct {
	Event        string
	ResourcePath string
	User         *users.User
}
