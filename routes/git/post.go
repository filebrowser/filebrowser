package git

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	s "github.com/hacdias/caddy-hugo/tools/server"
)

type postError struct {
	Message string `json:"message"`
}

// POST handles the POST method on GIT page which is only an API.
func POST(w http.ResponseWriter, r *http.Request) (int, error) {
	// Check if git is installed on the computer
	if _, err := exec.LookPath("git"); err != nil {
		return s.RespondJSON(w, &postError{"Git is not installed on your computer."}, 400, nil)
	}

	// Get the JSON information sent using a buffer
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buff.Bytes(), &info)

	// Check if command was sent
	if _, ok := info["command"]; !ok {
		return s.RespondJSON(w, &postError{"Command not specified."}, 400, nil)
	}

	command := info["command"].(string)
	args := strings.Split(command, " ")

	if len(args) > 0 && args[0] == "git" {
		args = append(args[:0], args[1:]...)
	}

	if len(args) == 0 {
		return s.RespondJSON(w, &postError{"Command not specified."}, 400, nil)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = conf.Path
	output, err := cmd.CombinedOutput()

	if err != nil {
		return s.RespondJSON(w, &postError{err.Error()}, 500, err)
	}

	return s.RespondJSON(w, &postError{string(output)}, 200, nil)
}
