package settings

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/hacdias/caddy-hugo/page"
	"github.com/spf13/hugo/parser"
)

// Execute the page
func Execute(w http.ResponseWriter, r *http.Request) (int, error) {
	if r.Method == "POST" {

	} else {
		frontmatter := getConfigFrontMatter()

		// 500 if the format of frontmatter can't be defined
		if frontmatter == "" {
			return 500, nil
		}

		config, err := getConfig(frontmatter)

		if err != nil {
			return 500, err
		}

		page := new(page.Page)
		page.Title = "settings"
		page.Body = config
		return page.Render("settings", w)
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

func getConfig(frontmatter string) (interface{}, error) {
	content := getConfigFileContent(frontmatter)

	switch frontmatter {
	case "yaml":
		return parser.HandleYAMLMetaData(content)
	case "json":
		return parser.HandleJSONMetaData(content)
	case "toml":
		return parser.HandleTOMLMetaData(content)
	}

	return []string{}, nil
}

func getConfigFileContent(frontmatter string) []byte {
	file, err := ioutil.ReadFile("config." + frontmatter)

	if err != nil {
		// there were a problem opening the file
		return []byte{}
	}

	return file
}
