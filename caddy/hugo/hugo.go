package hugo

import (
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
	Root string `description:"The relative or absolute path to the place where your website is located."`
	// Public folder
	Public string `description:"The relative or absolute path to the public folder."`
	// Hugo executable path
	Exe string `description:"The absolute path to the Hugo executable or the command to execute."`
	// Hugo arguments
	Args []string `description:"The arguments to run when running Hugo"`
	// Indicates if we should clean public before a new publish.
	CleanPublic bool `description:"Indicates if the public folder should be cleaned before publishing the website."`

	// TODO: admin interface to cgange options
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

	// If we are not using HTTP Post, we shall return Method Not Allowed
	// since we are only working with this method.
	if r.Method != http.MethodPost {
		return http.StatusMethodNotAllowed, nil
	}

	// If we are creating a file built from an archetype.
	if r.Header.Get("Archetype") != "" {
		if !c.User.AllowNew {
			return http.StatusForbidden, nil
		}

		filename := filepath.Join(string(c.User.FileSystem), r.URL.Path)
		archetype := r.Header.Get("archetype")

		ext := filepath.Ext(filename)

		// If the request isn't for a markdown file, we can't
		// handle it.
		if ext != ".markdown" && ext != ".md" {
			return http.StatusBadRequest, errUnsupportedFileType
		}

		// Tries to create a new file based on this archetype.
		args := []string{"new", filename, "--kind", archetype}
		if err := Run(h.Exe, args, h.Root); err != nil {
			return http.StatusInternalServerError, err
		}

		// Writes the location of the new file to the Header.
		w.Header().Set("Location", "/files/content/"+filename)
		return http.StatusCreated, nil
	}

	// If we are trying to regenerate the website.
	if r.Header.Get("Regenerate") == "true" {
		if !c.User.Permissions["allowPublish"] {
			return http.StatusForbidden, nil
		}

		filename := filepath.Join(string(c.User.FileSystem), r.URL.Path)

		// Before save command handler.
		if err := c.FM.Runner("before_publish", filename); err != nil {
			return http.StatusInternalServerError, err
		}

		// We only run undraft command if it is a file.
		if !strings.HasSuffix(filename, "/") {
			args := []string{"undraft", filename}
			if err := Run(h.Exe, args, h.Root); err != nil && !strings.Contains(err.Error(), "not a Draft") {
				return http.StatusInternalServerError, err
			}
		}

		// Regenerates the file
		h.run(false)

		// Executed the before publish command.
		if err := c.FM.Runner("before_publish", filename); err != nil {
			return http.StatusInternalServerError, err
		}

		return http.StatusOK, nil
	}

	if r.Header.Get("Schedule") != "" {
		if !c.User.Permissions["allowPublish"] {
			return http.StatusForbidden, nil
		}

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
