package git

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
)

// POST handles the POST method on GIT page which is only an API.
func POST(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Check if git is installed on the computer
	if _, err := exec.LookPath("git"); err != nil {
		jsonMessage(w, "Git is not installed on your computer.", 500)
		return 0, nil
	}

	// Get the JSON information sent using a buffer
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buff.Bytes(), &info)

	// Check if command was sent
	if _, ok := info["command"]; !ok {
		jsonMessage(w, "Command not specified.", 500)
		return 0, nil
	}

	command := info["command"].(string)
	args := strings.Split(command, " ")

	if len(args) > 0 && args[0] == "git" {
		args = append(args[:0], args[1:]...)
	}

	if len(args) == 0 {
		jsonMessage(w, "Command not specified.", 500)
		return 0, nil
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = c.Path
	output, err := cmd.CombinedOutput()

	if err != nil {
		jsonMessage(w, err.Error(), 500)
		return 0, nil
	}

	jsonMessage(w, string(output), 200)
	return 0, nil
}

type jsonMSG struct {
	Message string `json:"message"`
}

func jsonMessage(w http.ResponseWriter, message string, code int) {
	msg := &jsonMSG{
		Message: message,
	}

	m, _ := json.Marshal(msg)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(m)
}
