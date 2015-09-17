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

	/*



		objects := make([]interface{}, len(objectsNames))

		for index := range objectsNames {
			c := new(frontmatter)
			c.Type = "object"
			c.Title = objectsNames[index]

			if parent.Name == "" {
				c.Name = c.Title
			} else {
				c.Name = parent.Name + "[" + c.Name + "]"
			}

			c.Content = rawToPretty(config.(map[string]interface{})[c.Title], c)
			log.Print("\n\nObject Name:\n")
			log.Print(c.Name)
			objects[index] = c
		}

		arrays := make([]interface{}, len(arraysNames))

		for index := range arraysNames {
			c := new(frontmatter)
			c.Type = "array"
			c.Title = arraysNames[index]
			c.Name = parent.Name + c.Title + "[]"
			c.Content = rawToPretty(config.(map[string]interface{})[c.Title], c)
			log.Print("\n\nArray Name:\n")
			log.Print(c.Name)
			arrays[index] = c
		}

		/*strings := make([]interface{}, len(stringsNames))*/

	/*
		for index := range stringsNames {
			c := new(frontmatter)
			c.Title = stringsNames[index]
			c.Name = giveName(c.Title, parent)

			log.Print(c.Name)
		}

		/*	names := append(stringsNames, mapsNames...)

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
	*/
	//	settings := append(strings, slices..., maps...)

	/*if utils.IsSlice(config) {
		settings := make([]interface{}, len(config.([]interface{})))

		// TODO: improve this function

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

	*/
}
