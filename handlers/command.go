package handlers

import (
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/hacdias/caddy-filemanager/page"
)

// Command handles the requests for VCS related commands: git, svn and mercurial
func Command(w http.ResponseWriter, r *http.Request, c *config.Config, u *config.User) (int, error) {
	command := strings.Split(r.Header.Get("command"), " ")

	// Check if the command is allowed
	mayContinue := false

	for _, cmd := range u.Commands {
		if cmd == command[0] {
			mayContinue = true
		}
	}

	if !mayContinue {
		return http.StatusForbidden, nil
	}

	// Check if the program is talled is installed on the computer
	if _, err := exec.LookPath(command[0]); err != nil {
		return http.StatusNotImplemented, nil
	}

	path := strings.Replace(r.URL.Path, c.BaseURL, c.Scope, 1)
	path = filepath.Clean(path)

	cmd := exec.Command(command[0], command[1:len(command)]...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()

	if err != nil {
		return http.StatusInternalServerError, err
	}

	p := &page.Page{Info: &page.Info{Data: string(output)}}
	return p.PrintAsJSON(w)
}
