package frontmatter

import (
	"sort"

	"github.com/hacdias/caddy-hugo/utils"
	"github.com/spf13/hugo/parser"
)

// Pretty creates a new FrontMatter object
func Pretty(content []byte) (interface{}, error) {
	frontType := parser.DetectFrontMatter(rune(content[0]))
	front, err := frontType.Parse(content)

	if err != nil {
		return []string{}, err
	}

	return rawToPretty(front, "", ""), nil
}

type frontmatter struct {
	Name    string
	Content interface{}
	Parent  string
	Type    string
}

func rawToPretty(config interface{}, master string, parent string) interface{} {
	if utils.IsSlice(config) {
		settings := make([]interface{}, len(config.([]interface{})))

		for index, element := range config.([]interface{}) {
			c := new(frontmatter)
			c.Name = master
			c.Parent = parent

			if utils.IsMap(element) {
				c.Type = "object"
				c.Content = rawToPretty(element, c.Name, "object")
			} else if utils.IsSlice(element) {
				c.Type = "array"
				c.Content = rawToPretty(element, c.Name, "array")
			} else {
				c.Type = "text"
				c.Content = element
			}

			settings[index] = c
		}

		return settings
	}

	var mapsNames []string
	var stringsNames []string

	for index, element := range config.(map[string]interface{}) {
		if utils.IsMap(element) || utils.IsSlice(element) {
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
		c.Parent = parent

		i := config.(map[string]interface{})[names[index]]

		if utils.IsMap(i) {
			c.Type = "object"
			c.Content = rawToPretty(i, c.Name, "object")
		} else if utils.IsSlice(i) {
			c.Type = "array"
			c.Content = rawToPretty(i, c.Name, "array")
		} else {
			c.Type = "text"
			c.Content = i
		}

		settings[index] = c
	}

	return settings
}
