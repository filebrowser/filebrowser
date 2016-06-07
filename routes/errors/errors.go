package errors

import (
	"net/http"
	"strconv"
	"text/template"

	"github.com/hacdias/caddy-hugo/tools/server"
	"github.com/hacdias/caddy-hugo/tools/templates"
	"github.com/hacdias/caddy-hugo/tools/variables"
)

type errorInformation struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	err     error
}

// ServeHTTP is used to serve the content of GIT API.
func ServeHTTP(w http.ResponseWriter, r *http.Request, code int, err error) (int, error) {
	page := new(errorInformation)
	page.Title = strconv.Itoa(code) + " " + http.StatusText(code)
	page.err = err

	if err != nil {
		page.Message = err.Error()
	}

	switch r.Method {
	case "GET":
		functions := template.FuncMap{
			"Defined": variables.Defined,
		}

		var tpl *template.Template
		tpl, err = templates.Get(r, functions, "error")

		if err != nil {
			return http.StatusInternalServerError, err
		}

		err = tpl.Execute(w, page)

		if err != nil {
			return http.StatusInternalServerError, err
		}

		return 0, page.err
	default:
		return server.RespondJSON(w, page, code, err)
	}
}
