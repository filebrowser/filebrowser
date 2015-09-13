package edit

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hacdias/caddy-hugo/page"
)

type info struct {
	File string
}

// Execute sth
func Execute(w http.ResponseWriter, r *http.Request, file string) (int, error) {
	if r.Method == "POST" {
		// it's saving the post
	} else {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return 404, nil
		}

		file, err := ioutil.ReadFile(file)

		if err != nil {
			return 500, err
		}

		inf := new(info)
		inf.File = string(file)

		page := new(page.Page)
		page.Title = "Edit"
		page.Body = inf
		return page.Render("edit", w)
	}

	return 200, nil
}
