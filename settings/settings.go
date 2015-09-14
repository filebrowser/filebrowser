package settings

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/hacdias/caddy-hugo/frontmatter"
	"github.com/hacdias/caddy-hugo/page"
)

// Execute the page
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "POST" {

	} else {
		language := getConfigFrontMatter()

		if language == "" {
			log.Print("Configuration frontmatter can't be defined")
			return 500, nil
		}

		content, err := ioutil.ReadFile("config." + language)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		f, err := frontmatter.Pretty(content, language)

		if err != nil {
			log.Print(err)
			return 500, err
		}

		page := new(page.Page)
		page.Title = "Settings"
		page.Body = f
		return page.Render(w, "settings", "frontmatter")
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
