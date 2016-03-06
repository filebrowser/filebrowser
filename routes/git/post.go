package git

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/hacdias/caddy-hugo/config"
	"github.com/hacdias/caddy-hugo/tools/utils"
)

// POST handles the POST method on GIT page which is only an API.
func POST(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Check if git is installed on the computer
	if _, err := exec.LookPath("git"); err != nil {
		return utils.RespondJSON(w, map[string]string{
			"message": "Git is not installed on your computer.",
		}, 400, nil)
	}

	// Get the JSON information sent using a buffer
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buff.Bytes(), &info)

	// Check if command was sent
	if _, ok := info["command"]; !ok {
		return utils.RespondJSON(w, map[string]string{
			"message": "Command not specified.",
		}, 400, nil)
	}

	command := info["command"].(string)
	args := strings.Split(command, " ")

	if len(args) > 0 && args[0] == "git" {
		args = append(args[:0], args[1:]...)
	}

	if len(args) == 0 {
		return utils.RespondJSON(w, map[string]string{
			"message": "Command not specified.",
		}, 400, nil)
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = c.Path
	output, err := cmd.CombinedOutput()

	if err != nil {
		return utils.RespondJSON(w, map[string]string{
			"message": err.Error(),
		}, 500, err)
	}

	return utils.RespondJSON(w, map[string]string{
		"message": string(output),
	}, 200, nil)
}
