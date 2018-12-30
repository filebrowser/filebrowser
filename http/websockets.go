package http

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/filebrowser/filebrowser/search"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	cmdNotImplemented = []byte("Command not implemented.")
	cmdNotAllowed     = []byte("Command not allowed.")
)

func (e *Env) commandsHandler(w http.ResponseWriter, r *http.Request) {
	/* user, ok := e.getUser(w, r)
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
		message []byte
		command []string
	)

	// Starts an infinite loop until a valid command is captured.
	for {
		_, message, err = conn.ReadMessage()
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		command = strings.Split(string(message), " ")
		if len(command) != 0 {
			break
		}
	}

	allowed := false

	for _, cmd := range user.Commands {
		if regexp.MustCompile(cmd).MatchString(command[0]) {
			allowed = true
			break
		}
	}

	if !allowed {
		err = conn.WriteMessage(websocket.TextMessage, cmdNotAllowed)
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		}

		return
	}

	// Check if the program is installed on the computer.
	if _, err = exec.LookPath(command[0]); err != nil {
		err = conn.WriteMessage(websocket.TextMessage, cmdNotImplemented)
		if err != nil {
			httpErr(w, r, http.StatusInternalServerError, err)
			return
		} else {
			httpErr
		}

		return http.StatusNotImplemented, nil
	}

	// Gets the path and initializes a buffer.
	path := c.User.Scope + "/" + r.URL.Path
	path = filepath.Clean(path)
	buff := new(bytes.Buffer)

	// Sets up the command executation.
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = path
	cmd.Stderr = buff
	cmd.Stdout = buff

	// Starts the command and checks for fb.Errors.
	err = cmd.Start()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	// Set a 'done' variable to check whetever the command has already finished
	// running or not. This verification is done using a goroutine that uses the
	// method .Wait() from the command.
	done := false
	go func() {
		err = cmd.Wait()
		done = true
	}()

	// Function to print the current information on the buffer to the connection.
	print := func() error {
		by := buff.Bytes()
		if len(by) > 0 {
			err = conn.WriteMessage(websocket.TextMessage, by)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// While the command hasn't finished running, continue sending the output
	// to the client in intervals of 100 milliseconds.
	for !done {
		if err = print(); err != nil {
			return http.StatusInternalServerError, err
		}

		time.Sleep(100 * time.Millisecond)
	}

	// After the command is done executing, send the output one more time to the
	// browser to make sure it gets the latest information.
	if err = print(); err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil */

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
		if !user.IsAllowed(path) {
			return nil
		}

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
