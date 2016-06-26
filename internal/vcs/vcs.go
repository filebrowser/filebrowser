package vcs

import (
	"net/http"
	"os/exec"
	"strings"

	"github.com/hacdias/caddy-filemanager/internal/config"
	"github.com/hacdias/caddy-filemanager/internal/page"
)

// Handle handles the POST method on GIT page which is only an API.
func Handle(w http.ResponseWriter, r *http.Request, c *config.Config) (int, error) {
	command := strings.Split(r.Header.Get("command"), " ")

	// Check if the command is for git, mercurial or svn
	if command[0] != "git" && command[0] != "hg" && command[0] != "svn" {
		return http.StatusForbidden, nil
	}

	// Check if the program is talled is installed on the computer
	if _, err := exec.LookPath(command[0]); err != nil {
		return http.StatusNotImplemented, nil
	}

	cmd := exec.Command(command[0], command[1:len(command)]...)
	cmd.Dir = c.PathScope
	output, err := cmd.CombinedOutput()

	if err != nil {
		return http.StatusInternalServerError, err
	}

	page := &page.Page{Info: &page.Info{Data: string(output)}}
	return page.PrintAsJSON(w)
}
