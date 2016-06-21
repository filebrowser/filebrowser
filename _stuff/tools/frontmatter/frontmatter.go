package frontmatter

import (
	"log"
	"reflect"
	"sort"
	"strings"

	"github.com/hacdias/caddy-hugo/tools/variables"
	"github.com/spf13/cast"
	"github.com/spf13/hugo/parser"
)

const (
	mainName   = "#MAIN#"
	objectType = "object"
	arrayType  = "array"
)

var mainTitle = ""

// Pretty creates a new FrontMatter object
func Pretty(content []byte) (interface{}, string, error) {
	frontType := parser.DetectFrontMatter(rune(content[0]))
	front, err := frontType.Parse(content)

	if err != nil {
		return []string{}, mainTitle, err
	}

	object := new(frontmatter)
	object.Type = objectType
	object.Name = mainName

	return rawToPretty(front, object), mainTitle, nil
}

type frontmatter struct {
	Name     string
	Title    string
	Content  interface{}
	Type     string
	HTMLType string
	Parent   *frontmatter
}

func rawToPretty(config interface{}, parent *frontmatter) interface{} {
	objects := []*frontmatter{}
	arrays := []*frontmatter{}
	fields := []*frontmatter{}

	cnf := map[string]interface{}{}

	if reflect.TypeOf(config) == reflect.TypeOf(map[interface{}]interface{}{}) {
		for key, value := range config.(map[interface{}]interface{}) {
			cnf[key.(string)] = value
		}
	} else if reflect.TypeOf(config) == reflect.TypeOf([]interface{}{}) {
		for key, value := range config.([]interface{}) {
			cnf[string(key)] = value
		}
	} else {
		cnf = config.(map[string]interface{})
	}

	for name, element := range cnf {
		if variables.IsMap(element) {
			objects = append(objects, handleObjects(element, parent, name))
		} else if variables.IsSlice(element) {
			arrays = append(arrays, handleArrays(element, parent, name))
		} else {
			if name == "title" && parent.Name == mainName {
				mainTitle = element.(string)
			}

			fields = append(fields, handleFlatValues(element, parent, name))
		}
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
	c.Type = objectType
	c.Title = name

	if parent.Name == mainName {
		c.Name = c.Title
	} else if parent.Type == arrayType {
		c.Name = parent.Name + "[]"
	} else {
		c.Name = parent.Name + "[" + c.Title + "]"
	}

	c.Content = rawToPretty(content, c)
	return c
}

func handleArrays(content interface{}, parent *frontmatter, name string) *frontmatter {
	c := new(frontmatter)
	c.Parent = parent
	c.Type = arrayType
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

	c.Content = content

	switch strings.ToLower(name) {
	case "description":
		c.HTMLType = "textarea"
	case "date", "publishdate":
		c.HTMLType = "datetime"
		c.Content = cast.ToTime(content)
	default:
		c.HTMLType = "text"
	}

	if parent.Type == arrayType {
		c.Name = parent.Name + "[]"
		c.Title = content.(string)
	} else if parent.Type == objectType {
		c.Title = name
		c.Name = parent.Name + "[" + name + "]"

		if parent.Name == mainName {
			c.Name = name
		}
	} else {
		log.Panic("Parent type not allowed in handleFlatValues.")
	}

	return c
}
