package settings

import (
	"errors"
	"github.com/spf13/afero"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	invalidFilenameChars = regexp.MustCompile(`[^0-9A-Za-z@_\-.]`)

	dashes = regexp.MustCompile(`[\-]+`)
)

func CreateUserDir(username, userScope, serverRoot string, settings *Settings) (string, error) {
	var err error
	userScope = strings.TrimSpace(userScope)
	if userScope == "" || userScope == "./"  {
		userScope = "."
	}

	if !settings.CreateUserDir {
		return userScope, nil
	}

	fs := afero.NewBasePathFs(afero.NewOsFs(), serverRoot)

	//use the default auto create logic only if specific scope is not the default scope
	if userScope != settings.Defaults.Scope {
		//try create the dir, for example: settings.Defaults.Scope == "." and userScope == "./foo"
		if userScope != "." {
			err = fs.MkdirAll(userScope, os.ModePerm)
			if err != nil {
				log.Printf("create user: failed to mkdir user home dir: [%s]", userScope)
			}
		}
		return userScope, err
	}

	//clean username first
	username = cleanUsername(username)
	if username == "" || username == "-" || username == "." {
		log.Printf("create user: invalid user for home dir creation: [%s]", username)
		return "", errors.New("invalid user for home dir creation")
	}

	//create default user dir
	userHomeBase := settings.Defaults.Scope + string(os.PathSeparator) + "users"
	userHome := userHomeBase + string(os.PathSeparator) + username
	err = fs.MkdirAll(userHome, os.ModePerm)
	if err != nil {
		log.Printf("create user: failed to mkdir user home dir: [%s]", userHome)
	} else {
		log.Printf("create user: mkdir user home dir: [%s] successfully.", userHome)
	}
	return userHome,err
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