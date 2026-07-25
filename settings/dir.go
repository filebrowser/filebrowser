package settings

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/spf13/afero"

	"github.com/filebrowser/filebrowser/v2/users"
)

var (
	invalidFilenameChars = regexp.MustCompile(`[^0-9A-Za-z@_\-.]`)

	dashes = regexp.MustCompile(`[\-]+`)
)

// MakeUserDir makes the user directory according to settings.
func (s *Settings) MakeUserDir(username, userScope, serverRoot string) (string, error) {
	userScope = strings.TrimSpace(userScope)
	if userScope == "" && s.CreateUserDir {
		username = cleanUsername(username)
		if username == "" || username == "-" || username == "." {
			log.Printf("create user: invalid user for home dir creation: [%s]", username)
			return "", errors.New("invalid user for home dir creation")
		}
		userScope = path.Join(s.UserHomeBasePath, username)
	}

	userScope = path.Join("/", userScope)

	fs := afero.NewBasePathFs(afero.NewOsFs(), serverRoot)
	if err := fs.MkdirAll(userScope, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create user home dir: [%s]: %w", userScope, err)
	}
	return userScope, nil
}

// CreateUserHome derives and creates the home directory for a user that is
// being provisioned (via signup, proxy auth or hook auth) and sets user.Scope
// to the resulting path. When CreateUserDir is enabled and the caller did not
// supply an explicit scope, the scope is cleared so that MakeUserDir derives a
// per-user home from the username instead of falling back to the default scope
// (which normalizes to the server root, leaving every provisioned user sharing
// it).
//
// It reports whether the scope was derived from the username. A derived scope
// must be persisted with users.Storage.SaveProvisioned, which rejects a scope
// already owned by another user so that distinct usernames cannot silently
// share one home directory.
func (s *Settings) CreateUserHome(user *users.User, serverRoot string, explicitScope bool) (derived bool, err error) {
	derived = s.CreateUserDir && !explicitScope
	if derived {
		user.Scope = ""
	}

	userHome, err := s.MakeUserDir(user.Username, user.Scope, serverRoot)
	if err != nil {
		return false, err
	}
	user.Scope = userHome

	return derived, nil
}

func cleanUsername(s string) string {
	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")
	s = strings.ReplaceAll(s, "..", "")

	// Replace all characters which not in the list `0-9A-Za-z@_\-.` with a dash
	s = invalidFilenameChars.ReplaceAllString(s, "-")

	// Remove any multiple dashes caused by replacements above
	s = dashes.ReplaceAllString(s, "-")
	return s
}
