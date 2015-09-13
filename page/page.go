package page

import (
	"html/template"
	"net/http"
	"reflect"
	"strings"
	"unicode"

	"github.com/hacdias/caddy-hugo/assets"
)

const (
	templateExtension = ".tmpl"
	headerMark        = "{{#HEADER#}}"
	footerMark        = "{{#FOOTER#}}"
)

var funcMap = template.FuncMap{
	"splitCapitalize": splitCapitalize,
	"isMap":           isMap,
}

// TODO: utilspackage
func isMap(sth interface{}) bool {
	return reflect.ValueOf(sth).Kind() == reflect.Map
}

// TODO: utils package
func splitCapitalize(name string) string {
	var words []string
	l := 0
	for s := name; s != ""; s = s[l:] {
		l = strings.IndexFunc(s[1:], unicode.IsUpper) + 1
		if l <= 0 {
			l = len(s)
		}
		words = append(words, s[:l])
	}

	name = ""

	for _, element := range words {
		name += element + " "
	}

	name = strings.ToLower(name[:len(name)-1])
	name = strings.ToUpper(string(name[0])) + name[1:len(name)]

	return name
}

// Page type
type Page struct {
	Title string
	Body  interface{}
}

// Render the page
func (p *Page) Render(name string, w http.ResponseWriter) (int, error) {
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

	tpl, err := template.New("page").Funcs(funcMap).Parse(page)

	if err != nil {
		return 500, err
	}

	tpl.Execute(w, p)
	return 200, nil
}
