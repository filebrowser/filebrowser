package hugo

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
)

// HandleGit handles the POST method on GIT page which is only an API.
func HandleGit(w http.ResponseWriter, r *http.Request, c *Config) (int, error) {
	response := &Response{
		Code:    http.StatusOK,
		Err:     nil,
		Content: "OK",
	}

	// Check if git is installed on the computer
	if _, err := exec.LookPath("git"); err != nil {
		response.Code = http.StatusNotImplemented
		response.Content = "Git is not installed on your computer."
		return response.Send(w)
	}

	// Get the JSON information sent using a buffer
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buff.Bytes(), &info)

	// Check if command was sent
	if _, ok := info["command"]; !ok {
		response.Code = http.StatusBadRequest
		response.Content = "Command not specified."
		return response.Send(w)
	}

	command := info["command"].(string)
	args := strings.Split(command, " ")

	if len(args) > 0 && args[0] == "git" {
		args = append(args[:0], args[1:]...)
	}

	if len(args) == 0 {
		response.Code = http.StatusBadRequest
		response.Content = "Command not specified."
		return response.Send(w)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = c.Path
	output, err := cmd.CombinedOutput()

	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Content = err.Error()
		return response.Send(w)
	}

	response.Content = string(output)
	return response.Send(w)
}
