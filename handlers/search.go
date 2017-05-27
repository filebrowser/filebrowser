package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hacdias/caddy-filemanager/config"
)

type searchOptions struct {
	CaseInsensitive bool
	Terms           []string
}

func parseSearch(value string) *searchOptions {
	opts := &searchOptions{
		CaseInsensitive: strings.Contains(value, "case:insensitive"),
	}

	// removes the options from the value
	value = strings.Replace(value, "case:insensitive", "", -1)
	value = strings.Replace(value, "case:sensitive", "", -1)
	value = strings.TrimSpace(value)

	if opts.CaseInsensitive {
		value = strings.ToLower(value)
	}

	// if the value starts with " and finishes what that character, we will
	// only search for that term
	if value[0] == '"' && value[len(value)-1] == '"' {
		unique := strings.TrimPrefix(value, "\"")
		unique = strings.TrimSuffix(unique, "\"")

		opts.Terms = []string{unique}
		return opts
	}

	opts.Terms = strings.Split(value, " ")
	return opts
}

// Search ...
func Search(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
	// Upgrades the connection to a websocket and checks for errors.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var (
		value   string
		search  *searchOptions
		message []byte
	)

	// Starts an infinite loop until a valid command is captured.
	for {
		_, message, err = conn.ReadMessage()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if len(message) != 0 {
			value = string(message)
			break
		}
	}

	search = parseSearch(value)
	scope := strings.Replace(r.URL.Path, c.BaseURL, "", 1)
	scope = strings.TrimPrefix(scope, "/")
	scope = "/" + scope
	scope = u.Scope + scope
	scope = strings.Replace(scope, "\\", "/", -1)
	scope = filepath.Clean(scope)

	err = filepath.Walk(scope, func(path string, f os.FileInfo, err error) error {
		if search.CaseInsensitive {
			path = strings.ToLower(path)
		}

		path = strings.Replace(path, "\\", "/", -1)
		is := false

		for _, term := range search.Terms {
			if is {
				break
			}

			if strings.Contains(path, term) {
				if !u.Allowed(path) {
					return nil
				}

				is = true
			}
		}

		if !is {
			return nil
		}

		path = strings.TrimPrefix(path, scope)
		path = strings.TrimPrefix(path, "/")
		return conn.WriteMessage(websocket.TextMessage, []byte(path))
	})

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
