package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/hacdias/caddy-filemanager/config"
	"github.com/spf13/hugo/parser"
)

// PreProccessPUT is used to update a file that was edited
func PreProccessPUT(
	w http.ResponseWriter,
	r *http.Request,
	c *config.Config,
	u *config.User,
) (err error) {
	var (
		data      = map[string]interface{}{}
		file      []byte
		kind      string
		rawBuffer = new(bytes.Buffer)
	)

	kind = r.Header.Get("kind")
	rawBuffer.ReadFrom(r.Body)

	if kind != "" {
		err = json.Unmarshal(rawBuffer.Bytes(), &data)

		if err != nil {
			return
		}
	}

	switch kind {
	case "frontmatter-only":
		if file, err = ParseFrontMatterOnlyFile(data, r.URL.Path); err != nil {
			return
		}
	case "content-only":
		mainContent := data["content"].(string)
		mainContent = strings.TrimSpace(mainContent)
		file = []byte(mainContent)
	case "complete":
		if file, err = ParseCompleteFile(data, r.URL.Path, u.FrontMatter); err != nil {
			return
		}
	default:
		file = rawBuffer.Bytes()
	}

	// Overwrite the request Body
	r.Body = ioutil.NopCloser(bytes.NewReader(file))
	return
}

// ParseFrontMatterOnlyFile parses a frontmatter only file
func ParseFrontMatterOnlyFile(data interface{}, filename string) ([]byte, error) {
	frontmatter := strings.TrimPrefix(filepath.Ext(filename), ".")
	f, err := ParseFrontMatter(data, frontmatter)
	fString := string(f)

	// If it's toml or yaml, strip frontmatter identifier
	if frontmatter == "toml" {
		fString = strings.TrimSuffix(fString, "+++\n")
		fString = strings.TrimPrefix(fString, "+++\n")
	}

	if frontmatter == "yaml" {
		fString = strings.TrimSuffix(fString, "---\n")
		fString = strings.TrimPrefix(fString, "---\n")
	}

	f = []byte(fString)
	return f, err
}

// ParseFrontMatter is the frontmatter parser
func ParseFrontMatter(data interface{}, frontmatter string) ([]byte, error) {
	var mark rune

	switch frontmatter {
	case "toml":
		mark = rune('+')
	case "json":
		mark = rune('{')
	case "yaml":
		mark = rune('-')
	default:
		return []byte{}, errors.New("Can't define the frontmatter.")
	}

	f, err := parser.InterfaceToFrontMatter(data, mark)

	if err != nil {
		return []byte{}, err
	}

	return f, nil
}

// ParseCompleteFile parses a complete file
func ParseCompleteFile(data map[string]interface{}, filename string, frontmatter string) ([]byte, error) {
	mainContent := ""

	if _, ok := data["content"]; ok {
		// The main content of the file
		mainContent = data["content"].(string)
		mainContent = "\n\n" + strings.TrimSpace(mainContent) + "\n"

		// Removes the main content from the rest of the frontmatter
		delete(data, "content")
	}

	if _, ok := data["date"]; ok {
		data["date"] = data["date"].(string) + ":00"
	}

	front, err := ParseFrontMatter(data, frontmatter)

	if err != nil {
		fmt.Println(frontmatter)
		return []byte{}, err
	}

	// Generates the final file
	f := new(bytes.Buffer)
	f.Write(front)
	f.Write([]byte(mainContent))
	return f.Bytes(), nil
}
