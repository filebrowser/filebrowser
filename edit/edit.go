package edit

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/hacdias/caddy-hugo/frontmatter"
	"github.com/hacdias/caddy-hugo/page"
	"github.com/spf13/hugo/commands"
	"github.com/spf13/hugo/parser"
)

type information struct {
	Name        string
	Content     string
	FrontMatter interface{}
}

// Execute sth
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	filename := strings.Replace(r.URL.Path, "/admin/edit/", "", 1)

	if r.Method == "POST" {
		r.ParseForm()
		err := ioutil.WriteFile(filename, []byte(r.Form["content"][0]), 0666)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		commands.Execute()
	} else {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.Print(err)
			return 404, nil
		}

		reader, err := os.Open(filename)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		file, err := parser.ReadFrom(reader)

		inf := new(information)
		inf.Content = string(file.Content())
		inf.FrontMatter, err = frontmatter.Pretty(file.FrontMatter())

		if err != nil {
			log.Print(err)
			return 500, err
		}

		page := new(page.Page)
		page.Title = "Edit"
		page.Body = inf
		return page.Render(w, r, "edit", "frontmatter")
	}

	return 200, nil
}
