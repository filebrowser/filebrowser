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

	return rawToPretty(front, ""), nil
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
