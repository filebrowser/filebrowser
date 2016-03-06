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
	if _, err := exec.LookPath("git"); err != nil {
		jsonMessage(w, "Git is not installed on your computer.", false)
		return 200, nil
	}

	// Get the JSON information sent using a buffer
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buff.Bytes(), &info)

	if _, ok := info["command"]; !ok {
		jsonMessage(w, "Command not specified.", false)
		return 200, nil
	}

	command := info["command"].(string)
	args := strings.Split(command, " ")

	if len(args) > 0 && args[0] == "git" {
		args = append(args[:0], args[1:]...)
	}

	if len(args) == 0 {
		jsonMessage(w, "Command not specified.", false)
		return 200, nil
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = c.Path
	output, err := cmd.CombinedOutput()

	if err != nil {
		jsonMessage(w, err.Error(), false)
		return 200, nil
	}

	jsonMessage(w, string(output), true)
	return http.StatusOK, nil
}

type jsonMSG struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

func jsonMessage(w http.ResponseWriter, message string, success bool) {
	msg := &jsonMSG{
		Message: message,
		Success: success,
	}

	m, _ := json.Marshal(msg)
	w.Header().Set("Content-Type", "application/json")
	w.Write(m)
}
