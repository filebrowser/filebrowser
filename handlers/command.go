package handlers

import (
	"bytes"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hacdias/caddy-filemanager/config"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	cmdNotImplemented = []byte("Command not implemented.")
	cmdNotAllowed     = []byte("Command not allowed.")
)

// Command handles the requests for VCS related commands: git, svn and mercurial
func Command(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
	// Upgrades the connection to a websocket and checks for errors.
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return 0, err
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
			return http.StatusInternalServerError, err
		}

		command = strings.Split(string(message), " ")
		if len(command) != 0 {
			break
		}
	}

	// Check if the command is allowed
	allowed := false

	for _, cmd := range u.Commands {
		if cmd == command[0] {
			allowed = true
		}
	}

	if !allowed {
		err = conn.WriteMessage(websocket.BinaryMessage, cmdNotAllowed)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return 0, nil
	}

	// Check if the program is talled is installed on the computer.
	if _, err = exec.LookPath(command[0]); err != nil {
		err = conn.WriteMessage(websocket.BinaryMessage, cmdNotImplemented)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusNotImplemented, nil
	}

	// Gets the path and initializes a buffer.
	path := strings.Replace(r.URL.Path, c.BaseURL, c.Scope, 1)
	path = filepath.Clean(path)
	buff := new(bytes.Buffer)

	// Sets up the command executation.
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = path
	cmd.Stderr = buff
	cmd.Stdout = buff

	// Starts the command and checks for errors.
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

	return 0, nil
}
