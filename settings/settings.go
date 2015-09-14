package settings

import (
	"io/ioutil"
	"net/http"
	"os"
	"sort"

	"github.com/hacdias/caddy-hugo/page"
	"github.com/hacdias/caddy-hugo/utils"
	"github.com/spf13/hugo/parser"
)

type settings struct {
	Settings interface{}
	Keys     []string
}

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

		//	configIndex := getConfigNames(config)

		cnf := new(settings)
		cnf.Settings = getConfigNames(config, "")

		page := new(page.Page)
		page.Title = "Settings"
		page.Body = cnf
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

func getConfigFileContent(frontmatter string) []byte {
	file, err := ioutil.ReadFile("config." + frontmatter)

	if err != nil {
		// there were a problem opening the file
		return []byte{}
	}

	return file
}

// make it generic to frontmatter. everything bellow -> new file
func getConfig(frontmatter string) (interface{}, error) {
	content := getConfigFileContent(frontmatter)
	//	config := []string{}

	// get the config into a map
	if frontmatter == "yaml" {
		return parser.HandleYAMLMetaData(content)
	} else if frontmatter == "json" {
		return parser.HandleJSONMetaData(content)
	} else if frontmatter == "toml" {
		return parser.HandleTOMLMetaData(content)
	}

	return []string{}, nil
}

type conf struct {
	Name       string
	Master     string
	Content    interface{}
	SubContent bool
}

func getConfigNames(config interface{}, master string) interface{} {
	var mapsNames []string
	var stringsNames []string

	for index, element := range config.(map[string]interface{}) {
		if utils.IsMap(element) {
			mapsNames = append(mapsNames, index)
		} else {
			stringsNames = append(stringsNames, index)
		}
	}

	sort.Strings(mapsNames)
	sort.Strings(stringsNames)
	names := append(stringsNames, mapsNames...)

	settings := make([]interface{}, len(names))

	for index := range names {
		c := new(conf)
		c.Name = names[index]
		c.Master = master
		c.SubContent = false

		i := config.(map[string]interface{})[names[index]]

		if utils.IsMap(i) {
			c.Content = getConfigNames(i, c.Name)
			c.SubContent = true
		} else {
			c.Content = i
		}

		settings[index] = c
	}

	return settings
}
