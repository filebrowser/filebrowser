package frontmatter

import (
	"log"
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

	object := new(frontmatter)
	object.Type = "object"

	return rawToPretty(front, object), nil
}

type frontmatter struct {
	Name    string
	Title   string
	Content interface{}
	Type    string
	Parent  *frontmatter
}

func sortByTitle(config []*frontmatter) {
	keys := make([]string, len(config))
	positionByTitle := make(map[string]int)

	for index, element := range config {
		keys[index] = element.Title
		positionByTitle[element.Title] = index
	}

	sort.Strings(keys)
	cnf := make([]*frontmatter, len(config))

	for index, title := range keys {
		cnf[index] = config[positionByTitle[title]]
	}

	for index := range config {
		config[index] = cnf[index]
	}
}

func rawToPretty(config interface{}, parent *frontmatter) interface{} {
	objects := []*frontmatter{}
	arrays := []*frontmatter{}
	fields := []*frontmatter{}

	if parent.Type == "array" {
		for index, element := range config.([]interface{}) {
			c := new(frontmatter)
			c.Parent = parent

			if utils.IsMap(element) {
				c.Type = "object"

				if parent.Name == "" {
					c.Name = c.Title
				} else {
					c.Name = parent.Name + "[" + c.Name + "]"
				}

				c.Content = rawToPretty(config.([]interface{})[index], c)
				objects = append(objects, c)
			} else if utils.IsSlice(element) {
				c.Type = "array"
				c.Name = parent.Name + "[" + c.Name + "]"
				c.Content = rawToPretty(config.([]interface{})[index], c)

				arrays = append(arrays, c)
			} else {
				// TODO: add string, boolean, number
				c.Type = "string"
				c.Name = parent.Name + "[]"
				c.Title = element.(string)
				c.Content = config.([]interface{})[index]
				fields = append(fields, c)
			}
		}
	} else if parent.Type == "object" {
		for name, element := range config.(map[string]interface{}) {
			c := new(frontmatter)
			c.Title = name
			c.Parent = parent

			if utils.IsMap(element) {
				c.Type = "object"

				if parent.Name == "" {
					c.Name = c.Title
				} else {
					c.Name = parent.Name + "[" + c.Title + "]"
				}

				c.Content = rawToPretty(config.(map[string]interface{})[name], c)
				objects = append(objects, c)
			} else if utils.IsSlice(element) {
				c.Type = "array"

				if parent.Name == "" {
					c.Name = name
				} else {
					c.Name = parent.Name + "[" + c.Name + "]"
				}

				c.Content = rawToPretty(config.(map[string]interface{})[c.Title], c)

				arrays = append(arrays, c)
			} else {
				// TODO: add string, boolean, number
				c.Type = "string"

				if parent.Name == "" {
					c.Name = name
				} else {
					c.Name = parent.Name + "[" + name + "]"
				}

				c.Content = element
				fields = append(fields, c)
			}
		}
	} else {
		log.Panic("Parent type not allowed.")
	}

	sortByTitle(objects)
	sortByTitle(arrays)
	sortByTitle(fields)

	settings := []*frontmatter{}
	settings = append(settings, fields...)
	settings = append(settings, arrays...)
	settings = append(settings, objects...)

	return settings
}
