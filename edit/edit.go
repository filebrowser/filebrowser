package edit

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/page"
	"github.com/spf13/hugo/commands"
)

type fileInfo struct {
	Content string
	Name    string
}

// Execute sth
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	filename := strings.Replace(r.URL.Path, "admin/edit/", "", 1)

	if r.Method == "POST" {
		r.ParseForm()
		err := ioutil.WriteFile(filename, []byte(r.Form["content"][0]), 0666)

		if err != nil {
			return 500, err
		}

		commands.Execute()
	} else {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			return 404, nil
		}

		file, err := ioutil.ReadFile(filename)

		if err != nil {
			return 500, err
		}

		inf := new(fileInfo)
		inf.Content = string(file)
		inf.Name = filename

		page := new(page.Page)
		page.Title = "Edit"
		page.Body = inf
		return page.Render("edit", w)
	}

	return 200, nil
}
