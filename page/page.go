package page

import (
	"html/template"
	"log"
	"net/http"

	"github.com/hacdias/caddy-hugo/assets"
	"github.com/hacdias/caddy-hugo/utils"
)

const (
	templateExtension = ".tmpl"
)

var funcMap = template.FuncMap{
	"splitCapitalize": utils.SplitCapitalize,
}

// Page type
type Page struct {
	Title string
	Body  interface{}
}

// Render the page
func (p *Page) Render(w http.ResponseWriter, r *http.Request, templates ...string) (int, error) {
	if r.URL.Query().Get("minimal") == "true" {
		templates = append(templates, "base_minimal")
	} else {
		templates = append(templates, "base_full")
	}

	var tpl *template.Template

	for i, t := range templates {
		page, err := assets.Asset("templates/" + t + templateExtension)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		if i == 0 {
			tpl, err = template.New(t).Funcs(funcMap).Parse(string(page))
		} else {
			tpl, err = tpl.Parse(string(page))
		}

		if err != nil {
			log.Print(err)
			return 500, err
		}
	}

	tpl.Execute(w, p)
	return 200, nil
}
