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
func (p *Page) Render(name string, w http.ResponseWriter) (int, error) {
	base, err := assets.Asset("templates/base" + templateExtension)

	if err != nil {
		log.Print(err)
		return 500, err
	}

	page, err := assets.Asset("templates/" + name + templateExtension)

	if err != nil {
		log.Print(err)
		return 500, err
	}

	tpl, err := template.New("base").Funcs(funcMap).Parse(string(base))

	if err != nil {
		log.Print(err)
		return 500, err
	}

	tpl, err = tpl.Parse(string(page))

	if err != nil {
		log.Print(err)
		return 500, err
	}

	tpl.ExecuteTemplate(w, "base", p)
	return 200, nil
}
