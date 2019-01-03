package http

import (
	"bufio"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/filebrowser/filebrowser/search"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	cmdNotAllowed = []byte("Command not allowed.")
)

func (e *Env) commandsHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := e.getUser(w, r)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()

	var raw string

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
			return
		}

		raw = strings.TrimSpace(string(msg))
		if raw != "" {
			break
		}
	}

	if !user.CanExecute(strings.Split(raw, " ")[0]) {
		err := conn.WriteMessage(websocket.TextMessage, cmdNotAllowed)
		if err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
		}

		return
	}

	e.mux.RLock()
	command, err := e.Settings.ParseCommand(raw)
	e.mux.RUnlock()

	if err != nil {
		err := conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		if err != nil {
			wsErr(conn, r, http.StatusInternalServerError, err)
		}

		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/api/command")
	dir := user.FullPath(path)
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return
	}

	if err := cmd.Start(); err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
		return
	}

	s := bufio.NewScanner(io.MultiReader(stdout, stderr))
	for s.Scan() {
		conn.WriteMessage(websocket.TextMessage, s.Bytes())
	}

	if err := cmd.Wait(); err != nil {
		wsErr(conn, r, http.StatusInternalServerError, err)
	}
}

func (e *Env) searchHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := e.getUser(w, r)
	if !ok {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()

	var (
		value   string
		message []byte
	)

	for {
		_, message, err = conn.ReadMessage()
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		if len(message) != 0 {
			value = string(message)
			break
		}
	}

	scope := strings.TrimPrefix(r.URL.Path, "/api/search")
	err = search.Search(user.Fs, scope, value, func(path string, f os.FileInfo) error {
		response, _ := json.Marshal(map[string]interface{}{
			"dir":  f.IsDir(),
			"path": path,
		})

		return conn.WriteMessage(websocket.TextMessage, response)
	})

	if err != nil {
		httpErr(w, r, http.StatusInternalServerError, err)
		return
	}
}
