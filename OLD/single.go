package filemanager

import (
	"encoding/base64"
	"html"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"regexp"
)

var (
	videoRegex = regexp.MustCompile("video[/]")
	audioRegex = regexp.MustCompile("audio[/]")
	imageRegex = regexp.MustCompile("image[/]")
)

type File struct {
	*FileInfo
	Content string
}

// ServeSingleFile redirects the request for the respective method
func (f FileManager) ServeSingleFile(w http.ResponseWriter, r *http.Request, file *InfoRequest, c *Config) (int, error) {
	fullpath := c.PathScope + file.Path
	fullpath = filepath.Clean(fullpath)

	raw, err := ioutil.ReadFile(fullpath)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	base := base64.StdEncoding.EncodeToString(raw)
	mimetype := mime.TypeByExtension(filepath.Ext(file.Path))
	data := "data:" + mimetype + ";base64," + base

	page := &Page{
		Info: &PageInfo{
			Name: file.Path,
			Path: file.Path,
			Data: map[string]string{
				"Type":    RetrieveContentType(mimetype),
				"Base64":  data,
				"Content": html.EscapeString(string(raw)),
			},
		},
	}

	return page.PrintAsHTML(w, "single")
}

func RetrieveContentType(name string) string {
	if videoRegex.FindString(name) != "" {
		return "video"
	}

	if audioRegex.FindString(name) != "" {
		return "audio"
	}

	if imageRegex.FindString(name) != "" {
		return "image"
	}

	return "text"
}
