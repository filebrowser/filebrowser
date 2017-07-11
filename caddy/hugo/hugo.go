package hugo

import (
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	rice "github.com/GeertJohan/go.rice"
	"github.com/hacdias/filemanager"
	"github.com/hacdias/filemanager/variables"
	"github.com/robfig/cron"
)

type hugo struct {
	// Website root
	Root string
	// Public folder
	Public string
	// Hugo executable path
	Exe string
	// Hugo arguments
	Args []string
	// Indicates if we should clean public before a new publish.
	CleanPublic bool
	// A map of events to a slice of commands.
	Commands map[string][]string

	// AllowPublish

	javascript string
}

func (h hugo) BeforeAPI(c *filemanager.RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	// If we are using the 'magic url' for the settings, we should redirect the
	// request for the acutual path.
	if r.URL.Path == "/settings/" || r.URL.Path == "/settings" {
		var frontmatter string
		var err error

		if _, err = os.Stat(filepath.Join(h.Root, "config.yaml")); err == nil {
			frontmatter = "yaml"
		}

		if _, err = os.Stat(filepath.Join(h.Root, "config.json")); err == nil {
			frontmatter = "json"
		}

		if _, err = os.Stat(filepath.Join(h.Root, "config.toml")); err == nil {
			frontmatter = "toml"
		}

		r.URL.Path = "/config." + frontmatter
		return 0, nil
	}

	// From here on, we only care about 'hugo' router so we can bypass
	// the others.
	if c.Router != "hugo" {
		return 0, nil
	}

	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil
	}

	if r.Header.Get("Archetype") != "" {
		filename := filepath.Join(string(c.User.FileSystem), r.URL.Path)
		filename = filepath.Clean(filename)
		filename = strings.TrimPrefix(filename, "/")
		archetype := r.Header.Get("archetype")

		if !strings.HasSuffix(filename, ".md") && !strings.HasSuffix(filename, ".markdown") {
			return http.StatusBadRequest, errors.New("Your file must be markdown")
		}

		args := []string{"new", filename, "--kind", archetype}

		if err := Run(h.Exe, args, h.Root); err != nil {
			return http.StatusInternalServerError, err
		}

		w.Header().Set("Location", "/files/content/"+filename)
		return http.StatusCreated, nil
	}

	if r.Header.Get("Regenerate") == "true" {
		// Before save command handler.
		path := filepath.Clean(filepath.Join(string(c.User.FileSystem), r.URL.Path))
		if err := c.FM.Runner("before_publish", path); err != nil {
			return http.StatusInternalServerError, err
		}

		args := []string{"undraft", path}
		if err := Run(h.Exe, args, h.Root); err != nil {
			return http.StatusInternalServerError, err
		}

		h.run(false)

		if err := c.FM.Runner("before_publish", path); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	if r.Header.Get("Schedule") != "" {
		return h.schedule(c, w, r)
	}

	return http.StatusNotFound, nil
}

func (h hugo) AfterAPI(c *filemanager.RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	return 0, nil
}

func (h hugo) JavaScript() string {
	return rice.MustFindBox("./").MustString("hugo.js")
}

// run runs Hugo with the define arguments.
func (h hugo) run(force bool) {
	// If the CleanPublic option is enabled, clean it.
	if h.CleanPublic {
		os.RemoveAll(h.Public)
	}

	// Prevent running if watching is enabled
	if b, pos := variables.StringInSlice("--watch", h.Args); b && !force {
		if len(h.Args) > pos && h.Args[pos+1] != "false" {
			return
		}

		if len(h.Args) == pos+1 {
			return
		}
	}

	if err := Run(h.Exe, h.Args, h.Root); err != nil {
		log.Println(err)
	}
}

// schedule schedules a post to be published later.
func (h hugo) schedule(c *filemanager.RequestContext, w http.ResponseWriter, r *http.Request) (int, error) {
	t, err := time.Parse("2006-01-02T15:04", r.Header.Get("Schedule"))
	path := filepath.Join(string(c.User.FileSystem), r.URL.Path)
	path = filepath.Clean(path)

	if err != nil {
		return http.StatusInternalServerError, err
	}

	scheduler := cron.New()
	scheduler.AddFunc(t.Format("05 04 15 02 01 *"), func() {
		args := []string{"undraft", path}
		if err := Run(h.Exe, args, h.Root); err != nil {
			log.Printf(err.Error())
			return
		}

		h.run(false)
	})

	scheduler.Start()
	return http.StatusOK, nil
}
