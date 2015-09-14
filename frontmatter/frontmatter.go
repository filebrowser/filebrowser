package frontmatter

import (
	"sort"

	"github.com/hacdias/caddy-hugo/utils"
	"github.com/spf13/hugo/parser"
)

// Pretty creates a new FrontMatter object
func Pretty(content []byte, language string) (interface{}, error) {
	var err error
	var c interface{}

	if language == "yaml" {
		c, err = parser.HandleYAMLMetaData(content)
	} else if language == "json" {
		c, err = parser.HandleJSONMetaData(content)
	} else if language == "toml" {
		c, err = parser.HandleTOMLMetaData(content)
	}

	if err != nil {
		return []string{}, err
	}

	//log.Print(c)
	return rawToPretty(c, ""), nil
}

type frontmatter struct {
	Name       string
	Content    interface{}
	SubContent bool
}

func rawToPretty(config interface{}, master string) interface{} {
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
		c := new(frontmatter)
		c.Name = names[index]
		c.SubContent = false

		i := config.(map[string]interface{})[names[index]]

		if utils.IsMap(i) {
			c.Content = rawToPretty(i, c.Name)
			c.SubContent = true
		} else {
			c.Content = i
		}

		settings[index] = c
	}

	return settings
}
