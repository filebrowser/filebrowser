package main

import (
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/net/webdav"

	"github.com/hacdias/filemanager"
	handlers "github.com/hacdias/filemanager/http"
)

var cfg *filemanager.Config

func handler(w http.ResponseWriter, r *http.Request) {
	handlers.ServeHTTP(w, r, cfg)
}

func main() {
	cfg = &filemanager.Config{User: &filemanager.User{}}
	cfg.Scope = "."
	cfg.FileSystem = webdav.Dir(cfg.Scope)
	cfg.BaseURL = "/"
	cfg.HugoEnabled = false
	cfg.Users = map[string]*filemanager.User{}
	cfg.AllowCommands = true
	cfg.AllowEdit = true
	cfg.AllowNew = true
	cfg.Commands = []string{"git", "svn", "hg"}
	cfg.BeforeSave = func(r *http.Request, c *filemanager.Config, u *filemanager.User) error { return nil }
	cfg.AfterSave = func(r *http.Request, c *filemanager.Config, u *filemanager.User) error { return nil }
	cfg.Rules = []*filemanager.Rule{{
		Regex:  true,
		Allow:  false,
		Regexp: regexp.MustCompile("\\/\\..+"),
	}}

	cfg.BaseURL = strings.TrimPrefix(cfg.BaseURL, "/")
	cfg.BaseURL = strings.TrimSuffix(cfg.BaseURL, "/")
	cfg.BaseURL = "/" + cfg.BaseURL
	cfg.WebDavURL = ""

	if cfg.BaseURL == "/" {
		cfg.BaseURL = ""
	}

	if cfg.WebDavURL == "" {
		cfg.WebDavURL = "webdav"
	}

	cfg.PrefixURL = ""
	cfg.WebDavURL = cfg.BaseURL + "/" + strings.TrimPrefix(cfg.WebDavURL, "/")
	cfg.Handler = &webdav.Handler{
		Prefix:     cfg.WebDavURL,
		FileSystem: cfg.FileSystem,
		LockSystem: webdav.NewMemLS(),
	}

	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}