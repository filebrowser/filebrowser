package frontmatter

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/BurntSushi/toml"
	"github.com/hacdias/caddy-filemanager/utils/variables"

	"github.com/spf13/cast"
)

const (
	mainName   = "#MAIN#"
	objectType = "object"
	arrayType  = "array"
)

var mainTitle = ""

// Pretty creates a new FrontMatter object
func Pretty(content []byte) (*Content, string, error) {
	data, err := Unmarshal(content)

	if err != nil {
		return &Content{}, "", err
	}

	kind := reflect.ValueOf(data).Kind()

	if kind == reflect.Invalid {
		return &Content{}, "", nil
	}

	object := new(Block)
	object.Type = objectType
	object.Name = mainName

	if kind == reflect.Map {
		object.Type = objectType
	} else if kind == reflect.Slice || kind == reflect.Array {
		object.Type = arrayType
	}

	return rawToPretty(data, object), mainTitle, nil
}

// Unmarshal returns the data of the frontmatter
func Unmarshal(content []byte) (interface{}, error) {
	mark := rune(content[0])
	var data interface{}

	switch mark {
	case '-':
		// If it's YAML
		if err := yaml.Unmarshal(content, &data); err != nil {
			return nil, err
		}
	case '+':
		// If it's TOML
		content = bytes.Replace(content, []byte("+"), []byte(""), -1)
		if _, err := toml.Decode(string(content), &data); err != nil {
			return nil, err
		}
	case '{', '[':
		// If it's JSON
		if err := json.Unmarshal(content, &data); err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("Invalid frontmatter type")
	}

	return data, nil
}

// Marshal encodes the interface in a specific format
func Marshal(data interface{}, mark rune) ([]byte, error) {
	b := new(bytes.Buffer)

	switch mark {
	case '+':
		enc := toml.NewEncoder(b)
		err := enc.Encode(data)
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case '{':
		by, err := json.MarshalIndent(data, "", "   ")
		if err != nil {
			return nil, err
		}
		b.Write(by)
		_, err = b.Write([]byte("\n"))
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	case '-':
		by, err := yaml.Marshal(data)
		if err != nil {
			return nil, err
		}
		b.Write(by)
		_, err = b.Write([]byte("..."))
		if err != nil {
			return nil, err
		}
		return b.Bytes(), nil
	default:
		return nil, errors.New("Unsupported Format provided")
	}
}

// Content is the block content
type Content struct {
	Other   interface{}
	Fields  []*Block
	Arrays  []*Block
	Objects []*Block
}

// Block is a block
type Block struct {
	Name     string
	Title    string
	Type     string
	HTMLType string
	Content  *Content
	Parent   *Block
}

func rawToPretty(config interface{}, parent *Block) *Content {
	objects := []*Block{}
	arrays := []*Block{}
	fields := []*Block{}

	cnf := map[string]interface{}{}
	kind := reflect.TypeOf(config)

	switch kind {
	case reflect.TypeOf(map[interface{}]interface{}{}):
		for key, value := range config.(map[interface{}]interface{}) {
			cnf[key.(string)] = value
		}
	case reflect.TypeOf([]map[string]interface{}{}):
		for index, value := range config.([]map[string]interface{}) {
			cnf[strconv.Itoa(index)] = value
		}
	case reflect.TypeOf([]map[interface{}]interface{}{}):
		for index, value := range config.([]map[interface{}]interface{}) {
			cnf[strconv.Itoa(index)] = value
		}
	case reflect.TypeOf([]interface{}{}):
		for index, value := range config.([]interface{}) {
			cnf[strconv.Itoa(index)] = value
		}
	default:
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

	sort.Sort(sortByTitle(fields))
	sort.Sort(sortByTitle(arrays))
	sort.Sort(sortByTitle(objects))
	return &Content{
		Fields:  fields,
		Arrays:  arrays,
		Objects: objects,
	}
}

type sortByTitle []*Block

func (f sortByTitle) Len() int      { return len(f) }
func (f sortByTitle) Swap(i, j int) { f[i], f[j] = f[j], f[i] }
func (f sortByTitle) Less(i, j int) bool {
	return strings.ToLower(f[i].Name) < strings.ToLower(f[j].Name)
}

func handleObjects(content interface{}, parent *Block, name string) *Block {
	c := new(Block)
	c.Parent = parent
	c.Type = objectType
	c.Title = name

	if parent.Name == mainName {
		c.Name = c.Title
	} else if parent.Type == arrayType {
		c.Name = parent.Name + "[" + name + "]"
	} else {
		c.Name = parent.Name + "." + c.Title
	}

	c.Content = rawToPretty(content, c)
	return c
}

func handleArrays(content interface{}, parent *Block, name string) *Block {
	c := new(Block)
	c.Parent = parent
	c.Type = arrayType
	c.Title = name

	if parent.Name == mainName {
		c.Name = name
	} else {
		c.Name = parent.Name + "." + name
	}

	c.Content = rawToPretty(content, c)
	return c
}

func handleFlatValues(content interface{}, parent *Block, name string) *Block {
	c := new(Block)
	c.Parent = parent

	switch content.(type) {
	case bool:
		c.Type = "boolean"
	case int, float32, float64:
		c.Type = "number"
	default:
		c.Type = "string"
	}

	c.Content = &Content{Other: content}

	switch strings.ToLower(name) {
	case "description":
		c.HTMLType = "textarea"
	case "date", "publishdate":
		c.HTMLType = "datetime"
		c.Content = &Content{Other: cast.ToTime(content)}
	default:
		c.HTMLType = "text"
	}

	if parent.Type == arrayType {
		c.Name = parent.Name + "[]"
		c.Title = content.(string)
	} else if parent.Type == objectType {
		c.Title = name
		c.Name = parent.Name + "." + name

		if parent.Name == mainName {
			c.Name = name
		}
	} else {
		log.Panic("Parent type not allowed in handleFlatValues.")
	}

	return c
}
