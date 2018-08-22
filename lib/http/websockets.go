package http

import (
	"bytes"
	"encoding/json"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	fb "github.com/filebrowser/filebrowser/lib"
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

// command handles the requests for VCS related commands: git, svn and mercurial
func command(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Upgrades the connection to a websocket and checks for fb.Errors.
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

	for _, cmd := range c.User.Commands {
		if regexp.MustCompile(cmd).MatchString(command[0]) {
			allowed = true
			break
		}
	}

	if !allowed {
		err = conn.WriteMessage(websocket.TextMessage, cmdNotAllowed)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		return 0, nil
	}

	// Check if the program is installed on the computer.
	if _, err = exec.LookPath(command[0]); err != nil {
		err = conn.WriteMessage(websocket.TextMessage, cmdNotImplemented)
		if err != nil {
			return http.StatusInternalServerError, err
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

	return 0, nil
}

var (
	typeRegexp = regexp.MustCompile(`type:(\w+)`)
)

type condition func(path string) bool

type searchOptions struct {
	CaseSensitive bool
	Conditions    []condition
	Terms         []string
}

func extensionCondition(extension string) condition {
	return func(path string) bool {
		return filepath.Ext(path) == "."+extension
	}
}

func imageCondition(path string) bool {
	extension := filepath.Ext(path)
	mimetype := mime.TypeByExtension(extension)

	return strings.HasPrefix(mimetype, "image")
}

func audioCondition(path string) bool {
	extension := filepath.Ext(path)
	mimetype := mime.TypeByExtension(extension)

	return strings.HasPrefix(mimetype, "audio")
}

func videoCondition(path string) bool {
	extension := filepath.Ext(path)
	mimetype := mime.TypeByExtension(extension)

	return strings.HasPrefix(mimetype, "video")
}

func parseSearch(value string) *searchOptions {
	opts := &searchOptions{
		CaseSensitive: strings.Contains(value, "case:sensitive"),
		Conditions:    []condition{},
		Terms:         []string{},
	}

	// removes the options from the value
	value = strings.Replace(value, "case:insensitive", "", -1)
	value = strings.Replace(value, "case:sensitive", "", -1)
	value = strings.TrimSpace(value)

	types := typeRegexp.FindAllStringSubmatch(value, -1)
	for _, t := range types {
		if len(t) == 1 {
			continue
		}

		switch t[1] {
		case "image":
			opts.Conditions = append(opts.Conditions, imageCondition)
		case "audio", "music":
			opts.Conditions = append(opts.Conditions, audioCondition)
		case "video":
			opts.Conditions = append(opts.Conditions, videoCondition)
		default:
			opts.Conditions = append(opts.Conditions, extensionCondition(t[1]))
		}
	}

	if len(types) > 0 {
		// Remove the fields from the search value.
		value = typeRegexp.ReplaceAllString(value, "")
	}

	// If it's canse insensitive, put everything in lowercase.
	if !opts.CaseSensitive {
		value = strings.ToLower(value)
	}

	// Remove the spaces from the search value.
	value = strings.TrimSpace(value)

	if value == "" {
		return opts
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

// search searches for a file or directory.
func search(c *fb.Context, w http.ResponseWriter, r *http.Request) (int, error) {
	// Upgrades the connection to a websocket and checks for fb.Errors.
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
	scope := strings.TrimPrefix(r.URL.Path, "/")
	scope = "/" + scope
	scope = c.User.Scope + scope
	scope = strings.Replace(scope, "\\", "/", -1)
	scope = filepath.Clean(scope)

	err = filepath.Walk(scope, func(path string, f os.FileInfo, err error) error {
		var (
			originalPath string
		)

		path = strings.TrimPrefix(path, scope)
		path = strings.TrimPrefix(path, "/")
		path = strings.Replace(path, "\\", "/", -1)

		originalPath = path

		if !search.CaseSensitive {
			path = strings.ToLower(path)
		}

		// Only execute if there are conditions to meet.
		if len(search.Conditions) > 0 {
			match := false

			for _, t := range search.Conditions {
				if t(path) {
					match = true
					break
				}
			}

			// If doesn't meet the condition, go to the next.
			if !match {
				return nil
			}
		}

		if len(search.Terms) > 0 {
			is := false

			// Checks if matches the terms and if it is allowed.
			for _, term := range search.Terms {
				if is {
					break
				}

				if strings.Contains(path, term) {
					if !c.User.Allowed(path) {
						return nil
					}

					is = true
				}
			}

			if !is {
				return nil
			}
		}

		response, _ := json.Marshal(map[string]interface{}{
			"dir":  f.IsDir(),
			"path": originalPath,
		})

		return conn.WriteMessage(websocket.TextMessage, response)
	})

	if err != nil {
		return http.StatusInternalServerError, err
	}

	return 0, nil
}
