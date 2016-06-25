package git

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"

	"github.com/hacdias/caddy-filemanager/internal/config"
	"github.com/hacdias/caddy-filemanager/internal/page"
)

// Handle handles the POST method on GIT page which is only an API.
func Handle(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	// Check if git is installed on the computer
	if _, err := exec.LookPath("git"); err != nil {
		return http.StatusNotImplemented, nil
	}

	// Get the JSON information sent using a buffer
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.Body)

	// Creates the raw file "map" using the JSON
	var info map[string]interface{}
	json.Unmarshal(buff.Bytes(), &info)

	// Check if command was sent
	if _, ok := info["command"]; !ok {
		return http.StatusBadRequest, nil
	}

	command := info["command"].(string)
	args := strings.Split(command, " ")

	if len(args) > 0 && args[0] == "git" {
		args = append(args[:0], args[1:]...)
	}

	if len(args) == 0 {
		return http.StatusBadRequest, nil
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = c.PathScope
	output, err := cmd.CombinedOutput()

	if err != nil {
		return http.StatusInternalServerError, err
	}

	page := &page.Page{Info: &page.Info{Data: string(output)}}
	return page.PrintAsJSON(w)
}
