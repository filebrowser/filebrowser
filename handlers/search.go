package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/hacdias/caddy-filemanager/config"
)

// Search ...
func Search(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
	// Upgrades the connection to a websocket and checks for errors.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var (
		search  string
		message []byte
	)

	caseInsensitive := (r.URL.Query().Get("insensitive") == "true")

	// Starts an infinite loop until a valid command is captured.
	for {
		_, message, err = conn.ReadMessage()
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if len(message) != 0 {
			search = string(message)
			break
		}
	}

	if caseInsensitive {
		search = strings.ToLower(search)
	}

	scope := strings.Replace(r.URL.Path, c.BaseURL, "", 1)
	scope = strings.TrimPrefix(scope, "/")
	scope = "/" + scope
	scope = u.Scope + scope
	scope = strings.Replace(scope, "\\", "/", -1)
	scope = filepath.Clean(scope)

	err = filepath.Walk(scope, func(path string, f os.FileInfo, err error) error {
		if caseInsensitive {
			path = strings.ToLower(path)
		}

		if strings.Contains(path, search) {
			if !u.Allowed(path) {
				return nil
			}

			path = strings.TrimPrefix(path, scope)
			path = strings.Replace(path, "\\", "/", -1)
			path = strings.TrimPrefix(path, "/")

			err = conn.WriteMessage(websocket.TextMessage, []byte(path))
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}
