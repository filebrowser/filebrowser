package page

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/hacdias/caddy-hugo/assets"
)

const (
	templateExtension = ".tmpl"
	headerMark        = "{{#HEADER#}}"
	footerMark        = "{{#FOOTER#}}"
)

// Info type
type Info struct {
	Title string
	Body  interface{}
}

// Render the page
func (p *Info) Render(name string, w http.ResponseWriter) (int, error) {
	rawHeader, err := assets.Asset("templates/header" + templateExtension)

	if err != nil {
		return 500, err
	}

	header := string(rawHeader)

	rawFooter, err := assets.Asset("templates/footer" + templateExtension)

	if err != nil {
		return 500, err
	}

	footer := string(rawFooter)

	rawPage, err := assets.Asset("templates/" + name + templateExtension)

	if err != nil {
		return 500, err
	}

	page := string(rawPage)
	page = strings.Replace(page, headerMark, header, -1)
	page = strings.Replace(page, footerMark, footer, -1)

	tpl, err := template.New("page").Parse(page)

	if err != nil {
		return 500, err
	}

	tpl.Execute(w, p)
	return 200, nil
}
