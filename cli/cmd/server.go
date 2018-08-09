package cmd

import (
	"fmt"

	"github.com/spf13/viper"
	"github.com/asdine/storm"
	filebrowser "github.com/filebrowser/filebrowser/lib"
	"github.com/filebrowser/filebrowser/lib/bolt"
	h "github.com/filebrowser/filebrowser/lib/http"
	"github.com/filebrowser/filebrowser/lib/staticgen"
	"github.com/hacdias/fileutils"
	"gopkg.in/natefinch/lumberjack.v2"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
)

func Serve() {
	// Set up process log before anything bad happens.
	switch viper.GetString("Logger") {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(ioutil.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logfile,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}

	// Validate the provided config before moving forward
	if viper.GetString("AuthMethod") != "none" && viper.GetString("AuthMethod") != "default" && viper.GetString("AuthMethod") != "proxy" {
		log.Fatal("The property 'auth.method' needs to be set to 'default' or 'proxy'.")
	}

	if viper.GetString("AuthMethod") == "proxy" {
		if viper.GetString("LoginHeader") == "" {
			log.Fatal("The 'loginHeader' needs to be specified when 'proxy' authentication is used.")
		}
		log.Println("[WARN] Filebrowser authentication is configured to 'proxy' authentication. This can cause a huge security issue if the infrastructure is not configured correctly.")
	}

	// Builds the address and a listener.
	laddr := viper.GetString("Address") + ":" + viper.GetString("Port")
	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal(err)
	}

	// Tell the user the port in which is listening.
	fmt.Println("Listening on", listener.Addr().String())

	// Starts the server.
	if err := http.Serve(listener, handler()); err != nil {
		log.Fatal(err)
	}
}

func handler() http.Handler {
	db, err := storm.Open(viper.GetString("Database"))
	if err != nil {
		log.Fatal(err)
	}

	recaptchaHost := "https://www.google.com"
	if viper.GetBool("AlternativeRecaptcha") {
		recaptchaHost = "https://recaptcha.net"
	}

	fm := &filebrowser.FileBrowser{
		AuthMethod:      viper.GetString("AuthMethod"),
		LoginHeader:     viper.GetString("LoginHeader"),
		NoAuth:          viper.GetBool("NoAuth"),
		BaseURL:         viper.GetString("BaseURL"),
		PrefixURL:       viper.GetString("PrefixURL"),
		ReCaptchaHost:   recaptchaHost,
		ReCaptchaKey:    viper.GetString("ReCaptchaKey"),
		ReCaptchaSecret: viper.GetString("ReCaptchaSecret"),
		DefaultUser: &filebrowser.User{
			AllowCommands: viper.GetBool("AllowCommands"),
			AllowEdit:     viper.GetBool("AllowEdit"),
			AllowNew:      viper.GetBool("AllowNew"),
			AllowPublish:  viper.GetBool("AllowPublish"),
			Commands:      viper.GetStringSlice("Commands"),
			Rules:         []*filebrowser.Rule{},
			Locale:        viper.GetString("Locale"),
			CSS:           "",
			Scope:         viper.GetString("Scope"),
			FileSystem:    fileutils.Dir(viper.GetString("Scope")),
			ViewMode:      viper.GetString("ViewMode"),
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

	err = fm.Setup()
	if err != nil {
		log.Fatal(err)
	}

	switch viper.GetString("StaticGen") {
	case "hugo":
		hugo := &staticgen.Hugo{
			Root:        viper.GetString("Scope"),
			Public:      filepath.Join(viper.GetString("Scope"), "public"),
			Args:        []string{},
			CleanPublic: true,
		}

		if err = fm.Attach(hugo); err != nil {
			log.Fatal(err)
		}
	case "jekyll":
		jekyll := &staticgen.Jekyll{
			Root:        viper.GetString("Scope"),
			Public:      filepath.Join(viper.GetString("Scope"), "_site"),
			Args:        []string{"build"},
			CleanPublic: true,
		}

		if err = fm.Attach(jekyll); err != nil {
			log.Fatal(err)
		}
	}

	return h.Handler(fm)
}