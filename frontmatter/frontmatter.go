package frontmatter

import (
	"log"
	"reflect"
	"sort"

	"github.com/hacdias/caddy-hugo/utils"
	"github.com/spf13/hugo/parser"
)

const mainName = "#MAIN#"

// Pretty creates a new FrontMatter object
func Pretty(content []byte) (interface{}, error) {
	frontType := parser.DetectFrontMatter(rune(content[0]))
	front, err := frontType.Parse(content)

	if err != nil {
		return []string{}, err
	}

	object := new(frontmatter)
	object.Type = "object"
	object.Name = mainName

	return rawToPretty(front, object), nil
}

type frontmatter struct {
	Name    string
	Title   string
	Content interface{}
	Type    string
	Parent  *frontmatter
}

func rawToPretty(config interface{}, parent *frontmatter) interface{} {
	objects := []*frontmatter{}
	arrays := []*frontmatter{}
	fields := []*frontmatter{}

	if parent.Type == "array" {
		for index, element := range config.([]interface{}) {
			if utils.IsMap(element) {
				objects = append(objects, handleObjects(element, parent, string(index)))
			} else if utils.IsSlice(element) {
				arrays = append(arrays, handleArrays(element, parent, string(index)))
			} else {
				fields = append(fields, handleFlatValues(element, parent, string(index)))
			}
		}
	} else if parent.Type == "object" {
		for name, element := range config.(map[string]interface{}) {
			if utils.IsMap(element) {
				objects = append(objects, handleObjects(element, parent, name))
			} else if utils.IsSlice(element) {
				arrays = append(arrays, handleArrays(element, parent, name))
			} else {
				fields = append(fields, handleFlatValues(element, parent, name))
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

func sortByTitle(config []*frontmatter) {
	keys := make([]string, len(config))
	positionByTitle := make(map[string]int)

	for index, element := range config {
		keys[index] = element.Title
		positionByTitle[element.Title] = index
	}

	sort.Strings(keys)
	// TODO: http://golang.org/pkg/sort/#Interface
	cnf := make([]*frontmatter, len(config))

	for index, title := range keys {
		cnf[index] = config[positionByTitle[title]]
	}

	for index := range config {
		config[index] = cnf[index]
	}
}

func handleObjects(content interface{}, parent *frontmatter, name string) *frontmatter {
	c := new(frontmatter)
	c.Parent = parent
	c.Type = "object"
	c.Title = name

	if parent.Name == mainName {
		c.Name = c.Title
	} else {
		c.Name = parent.Name + "[" + c.Title + "]"
	}

	c.Content = rawToPretty(content, c)
	return c
}

func handleArrays(content interface{}, parent *frontmatter, name string) *frontmatter {
	c := new(frontmatter)
	c.Parent = parent
	c.Type = "array"
	c.Title = name

	if parent.Type == "object" && parent.Name == mainName {
		c.Name = name
	} else {
		c.Name = parent.Name + "[" + c.Name + "]"
	}

	c.Content = rawToPretty(content, c)
	return c
}

func handleFlatValues(content interface{}, parent *frontmatter, name string) *frontmatter {
	c := new(frontmatter)
	c.Parent = parent

	switch reflect.ValueOf(content).Kind() {
	case reflect.Bool:
		c.Type = "boolean"
	case reflect.Int:
	case reflect.Float32:
	case reflect.Float64:
		c.Type = "number"
	case reflect.String:
	default:
		c.Type = "string"
	}

	if parent.Type == "array" {
		c.Name = parent.Name + "[]"
		c.Title = content.(string)
	} else if parent.Type == "object" {
		c.Title = name
		c.Name = parent.Name + "[" + name + "]"

		if parent.Name == mainName {
			c.Name = name
		}
	} else {
		log.Panic("Parent type not allowed in handleFlatValues.")
	}

	c.Content = content
	return c
}
