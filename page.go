package filemanager

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)

// page contains the informations and functions needed to show the Page
type page struct {
	Info    *pageInfo
	Minimal bool
}

// pageInfo contains the information of a Page
type pageInfo struct {
	Name    string
	Path    string
	IsDir   bool
	User    *User
	Config  *Config
	Data    interface{}
	Editor  bool
	Display string
}

// BreadcrumbMapItem ...
type BreadcrumbMapItem struct {
	Name string
	URL  string
}

// BreadcrumbMap returns p.Path where every element is a map
// of URLs and path segment names.
func (i pageInfo) BreadcrumbMap() []BreadcrumbMapItem {
	result := []BreadcrumbMapItem{}

	if len(i.Path) == 0 {
		return result
	}

	// skip trailing slash
	lpath := i.Path
	if lpath[len(lpath)-1] == '/' {
		lpath = lpath[:len(lpath)-1]
	}

	parts := strings.Split(lpath, "/")
	for i, part := range parts {
		if i == len(parts)-1 {
			continue
		}

		if i == 0 && part == "" {
			result = append([]BreadcrumbMapItem{{
				Name: "/",
				URL:  "/",
			}}, result...)
			continue
		}

		result = append([]BreadcrumbMapItem{{
			Name: part,
			URL:  strings.Join(parts[:i+1], "/") + "/",
		}}, result...)
	}

	return result
}

// PreviousLink returns the path of the previous folder
func (i pageInfo) PreviousLink() string {
	path := strings.TrimSuffix(i.Path, "/")
	path = strings.TrimPrefix(path, "/")
	path = i.Config.AbsoluteURL() + "/" + path
	path = path[0 : len(path)-len(i.Name)]

	if len(path) < len(i.Config.AbsoluteURL()+"/") {
		return ""
	}

	return path
}

// PrintHTML formats the page in HTML and executes the template
func (p page) PrintHTML(w http.ResponseWriter, templates ...string) (int, error) {

	if p.Minimal {
		templates = append(templates, "minimal")
	} else {
		templates = append(templates, "base")
	}

	var tpl *template.Template

	// For each template, add it to the the tpl variable
	for i, t := range templates {
		// Get the template from the assets
		//Page, err := assets.Asset("templates/" + t + ".tmpl")
		Page, err := []byte("Shit"), errors.New("Hello")

		// Check if there is some error. If so, the template doesn't exist
		if err != nil {
			log.Print(err)
			return http.StatusInternalServerError, err
		}

		// If it's the first iteration, creates a new template and add the
		// functions map
		if i == 0 {
			tpl, err = template.New(t).Funcs(functionMap).Parse(string(Page))
		} else {
			tpl, err = tpl.Parse(string(Page))
		}

		if err != nil {
			log.Print(err)
			return http.StatusInternalServerError, err
		}
	}

	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, p.Info)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, err = buf.WriteTo(w)
	return http.StatusOK, err
}

// PrintJSON prints the current Page information in JSON
func (p page) PrintJSON(w http.ResponseWriter) (int, error) {
	marsh, err := json.MarshalIndent(p.Info.Data, "", "    ")
	if err != nil {
		return http.StatusInternalServerError, err
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if _, err := w.Write(marsh); err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusOK, nil
}

// printError prints the error page
func printError(w http.ResponseWriter, code int, err error) (int, error) {
	tpl := errorTemplate
	tpl = strings.Replace(tpl, "TITLE", strconv.Itoa(code)+" "+http.StatusText(code), -1)
	tpl = strings.Replace(tpl, "CODE", err.Error(), -1)

	_, err = w.Write([]byte(tpl))

	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, nil
}

const errorTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>TITLE</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta charset="utf-8">
    <style>
    html {
        background-color: #2196f3;
        color: #fff;
        font-family: sans-serif;
    }
    code {
        background-color: rgba(0,0,0,0.1);
        border-radius: 5px;
        padding: 1em;
        display: block;
        box-sizing: border-box;
    }
    .center {
        max-width: 40em;
        margin: 2em auto 0;
    }
    a {
        text-decoration: none;
        color: #eee;
        font-weight: bold;
    }
	p {
		line-height: 1.3;
	}
    </style>
</head>

<body>
    <div class="center">
        <h1>TITLE</h1>

        <p>Try reloading the page or hitting the back button. If this error persists, it seems that you may have found a bug! Please create an issue at <a href="https://github.com/hacdias/caddy-filemanager/issues">hacdias/caddy-filemanager</a> repository on GitHub with the code below.</p>

        <code>CODE</code>
    </div>
</html>`
