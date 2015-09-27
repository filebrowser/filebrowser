package frontmatter

import (
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/hacdias/caddy-cms/utils"
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

	sort.Sort(sortByTitle(objects))
	sort.Sort(sortByTitle(arrays))
	sort.Sort(sortByTitle(fields))

	settings := []*frontmatter{}
	settings = append(settings, fields...)
	settings = append(settings, arrays...)
	settings = append(settings, objects...)
	return settings
}

type sortByTitle []*frontmatter

func (f sortByTitle) Len() int      { return len(f) }
func (f sortByTitle) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f sortByTitle) Less(i, j int) bool {
	return strings.ToLower(f[i].Name) < strings.ToLower(f[j].Name)
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

	if parent.Name == mainName {
		c.Name = name
	} else {
		c.Name = parent.Name + "[" + name + "]"
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
	case reflect.Int, reflect.Float32, reflect.Float64:
		c.Type = "number"
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
