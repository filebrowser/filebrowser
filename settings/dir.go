package settings

import (
	"errors"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/afero"
)

var (
	invalidFilenameChars = regexp.MustCompile(`[^0-9A-Za-z@_\-.]`)

	dashes = regexp.MustCompile(`[\-]+`)
)

// MakeUserDir makes the user directory according to settings.
func (s *Settings) MakeUserDir(username, userScope, serverRoot string) (string, error) {
	var err error
	userScope = strings.TrimSpace(userScope)
	if userScope == "" || userScope == "./" {
		userScope = "."
	}

	if !s.CreateUserDir {
		return userScope, nil
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), serverRoot)

	// Use the default auto create logic only if specific scope is not the default scope
	if userScope != s.Defaults.Scope {
		// Try create the dir, for example: settings.Defaults.Scope == "." and userScope == "./foo"
		if userScope != "." {
			err = fs.MkdirAll(userScope, os.ModePerm)
			if err != nil {
				log.Printf("create user: failed to mkdir user home dir: [%s]", userScope)
			}
		}
		return userScope, err
	}

	// Clean username first
	username = cleanUsername(username)
	if username == "" || username == "-" || username == "." {
		log.Printf("create user: invalid user for home dir creation: [%s]", username)
		return "", errors.New("invalid user for home dir creation")
	}

	// Create default user dir
	userHomeBase := s.Defaults.Scope + string(os.PathSeparator) + "users"
	userHome := userHomeBase + string(os.PathSeparator) + username
	err = fs.MkdirAll(userHome, os.ModePerm)
	if err != nil {
		log.Printf("create user: failed to mkdir user home dir: [%s]", userHome)
	} else {
		log.Printf("create user: mkdir user home dir: [%s] successfully.", userHome)
	}
	return userHome, err
}

func cleanUsername(s string) string {
	// Remove any trailing space to avoid ending on -
	s = strings.Trim(s, " ")
	s = strings.Replace(s, "..", "", -1)

	// Replace all characters which not in the list `0-9A-Za-z@_\-.` with a dash
	s = invalidFilenameChars.ReplaceAllString(s, "-")

	// Remove any multiple dashes caused by replacements above
	s = dashes.ReplaceAllString(s, "-")
	return s
}
