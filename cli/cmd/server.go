package cmd

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"

	"github.com/asdine/storm"
	filebrowser "github.com/filebrowser/filebrowser/lib"
	"github.com/filebrowser/filebrowser/lib/bolt"
	h "github.com/filebrowser/filebrowser/lib/http"
	"github.com/filebrowser/filebrowser/lib/staticgen"
	"github.com/hacdias/fileutils"
	"github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

func Serve() {
	// Set up process log before anything bad happens.
	switch l := viper.GetString("log"); l {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   l,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}

	// Validate the provided config before moving forward
	{
		// Map of valid authentication methods, containing a boolean value to indicate the need of Auth.Header
		validMethods := make(map[string]bool)
		validMethods["none"] = false
		validMethods["default"] = false
		validMethods["proxy"] = true

		m := viper.GetString("auth.method")
		b, ok := validMethods[m]
		if !ok {
			log.Fatal("The property 'auth.method' needs to be set to 'none', 'default' or 'proxy'.")
		}

		if b {
			if viper.GetString("auth.header") == "" {
				log.Fatal("The 'auth.header' needs to be specified when '", m, "' authentication is used.")
			}
			log.Println("[WARN] Filebrowser authentication is configured to '", m, "' authentication. This can cause a huge security issue if the infrastructure is not configured correctly.")
		}
	}

	// Builds the address and a listener.
	laddr := viper.GetString("address") + ":" + viper.GetString("port")
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	// Tell the user the port in which is listening.
	log.Println("Listening on", listener.Addr().String())

	// Starts the server.
	if err := http.Serve(listener, handler()); err != nil {
		log.Fatal(err)
	}
}

func handler() http.Handler {
	db, err := storm.Open(viper.GetString("database"))
	if err != nil {
		log.Fatal(err)
	}

	fb := &filebrowser.FileBrowser{
		Auth: &filebrowser.Auth{
			Method: viper.GetString("auth.method"),
			Header: viper.GetString("auth.header"),
		},
		ReCaptcha: &filebrowser.ReCaptcha{
			Host:   viper.GetString("recaptcha.host"),
			Key:    viper.GetString("recaptcha.key"),
			Secret: viper.GetString("recaptcha.secret"),
		},
		DefaultUser: &filebrowser.User{
			AllowCommands: viper.GetBool("defaults.allowCommands"),
			AllowEdit:     viper.GetBool("defaults.allowEdit"),
			AllowNew:      viper.GetBool("defaults.allowNew"),
			AllowPublish:  viper.GetBool("defaults.allowPublish"),
			Commands:      viper.GetStringSlice("defaults.commands"),
			Rules:         []*filebrowser.Rule{},
			Locale:        viper.GetString("defaults.locale"),
			CSS:           "",
			Scope:         viper.GetString("defaults.scope"),
			FileSystem:    fileutils.Dir(viper.GetString("defaults.scope")),
			ViewMode:      viper.GetString("defaults.viewMode"),
		},
		Store: &filebrowser.Store{
			Config: bolt.ConfigStore{DB: db},
			Users:  bolt.UsersStore{DB: db},
			Share:  bolt.ShareStore{DB: db},
		},
		NewFS: func(scope string) filebrowser.FileSystem {
			return fileutils.Dir(scope)
		},
	}

	fb.SetBaseURL(viper.GetString("baseurl"))
	fb.SetPrefixURL(viper.GetString("prefixurl"))

	err = fb.Setup()
	if err != nil {
		log.Fatal(err)
	}

	switch viper.GetString("staticgen") {
	case "hugo":
		hugo := &staticgen.Hugo{
			Root:        viper.GetString("Scope"),
			Public:      filepath.Join(viper.GetString("Scope"), "public"),
			Args:        []string{},
			CleanPublic: true,
		}

		if err = fb.Attach(hugo); err != nil {
			log.Fatal(err)
		}
	case "jekyll":
		jekyll := &staticgen.Jekyll{
			Root:        viper.GetString("Scope"),
			Public:      filepath.Join(viper.GetString("Scope"), "_site"),
			Args:        []string{"build"},
			CleanPublic: true,
		}

		if err = fb.Attach(jekyll); err != nil {
			log.Fatal(err)
		}
	}

	return h.Handler(fb)
}
