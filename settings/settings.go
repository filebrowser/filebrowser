package settings

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"

	"github.com/hacdias/caddy-hugo/frontmatter"
	"github.com/hacdias/caddy-hugo/utils"
)

type page struct {
	Name     string
	Settings interface{}
}

// Execute the page
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	language := getConfigFrontMatter()

	if language == "" {
		log.Print("Configuration frontmatter can't be defined")
		return 500, nil
	}

	if r.Method == "POST" {
		err := os.Remove("config." + language)

		if err != nil {
			log.Print(err)
			return 500, nil
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		raw := buf.Bytes()

		content := new(bytes.Buffer)
		json.Indent(content, raw, "", "  ")

		err = ioutil.WriteFile("config.json", content.Bytes(), 0666)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{}"))
	} else {
		content, err := ioutil.ReadFile("config." + language)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		f, err := frontmatter.Pretty(appendFrontMatterRune(content, language))

		if err != nil {
			log.Print(err)
			return 500, err
		}

		functions := template.FuncMap{
			"splitCapitalize": utils.SplitCapitalize,
		}

		tpl, err := utils.GetTemplate(r, functions, "settings", "frontmatter")

		if err != nil {
			log.Print(err)
			return 500, err
		}

		p := new(page)
		p.Name = "settings"
		p.Settings = f

		tpl.Execute(w, p)
	}
	return 200, nil
}

func getConfigFrontMatter() string {
	var frontmatter string

	if _, err := os.Stat("config.yaml"); err == nil {
		frontmatter = "yaml"
	}

	if _, err := os.Stat("config.json"); err == nil {
		frontmatter = "json"
	}

	if _, err := os.Stat("config.toml"); err == nil {
		frontmatter = "toml"
	}

	return frontmatter
}

func appendFrontMatterRune(frontmatter []byte, language string) []byte {
	switch language {
	case "yaml":
		return []byte("---\n" + string(frontmatter) + "\n---")
	case "toml":
		return []byte("+++\n" + string(frontmatter) + "\n+++")
	case "json":
		return frontmatter
	}

	return frontmatter
}
