package cmd

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	v "github.com/spf13/viper"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/diskcache"
	fbErrors "github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/frontend"
	fbhttp "github.com/filebrowser/filebrowser/v2/http"
	"github.com/filebrowser/filebrowser/v2/img"
	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/storage"
	"github.com/filebrowser/filebrowser/v2/users"
)

var (
	cfgFile string

	flatNamesMigrations = map[string]string{
		"file-mode":                        "fileMode",
		"dir-mode":                         "dirMode",
		"hide-login-button":                "hideLoginButton",
		"create-user-dir":                  "createUserDir",
		"minimum-password-length":          "minimumPasswordLength",
		"socket-perm":                      "socketPerm",
		"disable-thumbnails":               "disableThumbnails",
		"disable-preview-resize":           "disablePreviewResize",
		"disable-exec":                     "disableExec",
		"disable-type-detection-by-header": "disableTypeDetectionByHeader",
		"img-processors":                   "imageProcessors",
		"cache-dir":                        "cacheDir",
		"token-expiration-time":            "tokenExpirationTime",
		"baseurl":                          "baseURL",
	}
)

func migrateFlagNames(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if newName, ok := flatNamesMigrations[name]; ok {
		name = newName
	}

	return pflag.NormalizedName(name)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.SilenceUsage = true
	cobra.MousetrapHelpText = ""

	rootCmd.SetVersionTemplate("File Browser version {{printf \"%s\" .Version}}\n")

	flags := rootCmd.Flags()
	persistent := rootCmd.PersistentFlags()

	persistent.StringVarP(&cfgFile, "config", "c", "", "config file path")
	persistent.StringP("database", "d", "./filebrowser.db", "database path")
	flags.Bool("noauth", false, "use the noauth auther when using quick setup")
	flags.String("username", "admin", "username for the first user when using quick config")
	flags.String("password", "", "hashed password for the first user when using quick config")

	addServerFlags(flags)

	rootCmd.SetGlobalNormalizationFunc(migrateFlagNames)
}

func addServerFlags(flags *pflag.FlagSet) {
	flags.StringP("address", "a", "127.0.0.1", "address to listen on")
	flags.StringP("log", "l", "stdout", "log output")
	flags.StringP("port", "p", "8080", "port to listen on")
	flags.StringP("cert", "t", "", "tls certificate")
	flags.StringP("key", "k", "", "tls key")
	flags.StringP("root", "r", ".", "root to prepend to relative paths")
	flags.String("socket", "", "socket to listen to (cannot be used with address, port, cert nor key flags)")
	flags.Uint32("socketPerm", 0666, "unix socket file permissions")
	flags.StringP("baseURL", "b", "", "base url")
	flags.String("cacheDir", "", "file cache directory (disabled if empty)")
	flags.String("tokenExpirationTime", "2h", "user session timeout")
	flags.Int("imageProcessors", 4, "image processors count")
	flags.Bool("disableThumbnails", false, "disable image thumbnails")
	flags.Bool("disablePreviewResize", false, "disable resize of image previews")
	flags.Bool("disableExec", true, "disables Command Runner feature")
	flags.Bool("disableTypeDetectionByHeader", false, "disables type detection by reading file headers")
}

var rootCmd = &cobra.Command{
	Use:   "filebrowser",
	Short: "A stylish web-based file browser",
	Long: `File Browser CLI lets you create the database to use with File Browser,
manage your users and all the configurations without accessing the
web interface.

If you've never run File Browser, you'll need to have a database for
it. Don't worry: you don't need to setup a separate database server.
We're using Bolt DB which is a single file database and all managed
by ourselves.

All the flags you have available (except "config" for the configuration file),
can be given either through environment variables or configuration files.

If you don't set "config", it will look for a configuration file called
.filebrowser.{json, toml, yaml, yml} in the following directories:

- ./
- $HOME/
- /etc/filebrowser/

The precedence of the configuration values are as follows:

- flags
- environment variables
- configuration file
- database values
- defaults

The environment variables are prefixed by "FB_" followed by the option
name in caps. So to set "database" via an env variable, you should
set FB_DATABASE.

Also, if the database path doesn't exist, File Browser will enter into
the quick setup mode and a new database will be bootstrapped and a new
user created with the credentials from options "username" and "password".`,
	RunE: python(func(cmd *cobra.Command, _ []string, d *pythonData) error {
		log.Println(cfgFile)

		if !d.hadDB {
			err := quickSetup(*d)
			if err != nil {
				return err
			}
		}

		// build img service
		workersCount := v.GetInt("imageprocessors")
		if workersCount < 1 {
			return errors.New("image resize workers count could not be < 1")
		}
		imgSvc := img.New(workersCount)

		var fileCache diskcache.Interface = diskcache.NewNoOp()
		cacheDir, err := cmd.Flags().GetString("cachedir")
		if err != nil {
			return err
		}
		if cacheDir != "" {
			if err := os.MkdirAll(cacheDir, 0700); err != nil {
				return fmt.Errorf("can't make directory %s: %w", cacheDir, err)
			}
			fileCache = diskcache.New(afero.NewOsFs(), cacheDir)
		}

		server, err := getRunParams(d.store)
		if err != nil {
			return err
		}
		setupLog(server.Log)

		root, err := filepath.Abs(server.Root)
		if err != nil {
			return err
		}
		server.Root = root

		adr := server.Address + ":" + server.Port

		var listener net.Listener

		switch {
		case server.Socket != "":
			listener, err = net.Listen("unix", server.Socket)
			if err != nil {
				return err
			}
			socketPerm, err := cmd.Flags().GetUint32("socketperm")
			if err != nil {
				return err
			}
			err = os.Chmod(server.Socket, os.FileMode(socketPerm))
			if err != nil {
				return err
			}
		case server.TLSKey != "" && server.TLSCert != "":
			cer, err := tls.LoadX509KeyPair(server.TLSCert, server.TLSKey)
			if err != nil {
				return err
			}
			listener, err = tls.Listen("tcp", adr, &tls.Config{
				MinVersion:   tls.VersionTLS12,
				Certificates: []tls.Certificate{cer}},
			)
			if err != nil {
				return err
			}
		default:
			listener, err = net.Listen("tcp", adr)
			if err != nil {
				return err
			}
		}

		assetsFs, err := fs.Sub(frontend.Assets(), "dist")
		if err != nil {
			panic(err)
		}

		handler, err := fbhttp.NewHandler(imgSvc, fileCache, d.store, server, assetsFs)
		if err != nil {
			return err
		}

		defer listener.Close()

		log.Println("Listening on", listener.Addr().String())
		srv := &http.Server{
			Handler:           handler,
			ReadHeaderTimeout: 60 * time.Second,
		}

		go func() {
			if err := srv.Serve(listener); !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("HTTP server error: %v", err)
			}

			log.Println("Stopped serving new connections.")
		}()

		sigc := make(chan os.Signal, 1)
		signal.Notify(sigc,
			os.Interrupt,
			syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)
		sig := <-sigc
		log.Println("Got signal:", sig)

		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownRelease()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("HTTP shutdown error: %v", err)
		}
		log.Println("Graceful shutdown complete.")

		switch sig {
		case syscall.SIGHUP:
			d.err = fbErrors.ErrSighup
		case syscall.SIGINT:
			d.err = fbErrors.ErrSigint
		case syscall.SIGQUIT:
			d.err = fbErrors.ErrSigquit
		case syscall.SIGTERM:
			d.err = fbErrors.ErrSigTerm
		}

		return d.err
	}, pythonConfig{allowNoDB: true}),
}

func getRunParams(st *storage.Storage) (*settings.Server, error) {
	server, err := st.Settings.GetServer()
	if err != nil {
		return nil, err
	}

	if val, set := getStringParamB("root"); set {
		server.Root = val
	}

	if val, set := getStringParamB("baseurl"); set {
		server.BaseURL = val
	}

	if val, set := getStringParamB("log"); set {
		server.Log = val
	}

	isSocketSet := false
	isAddrSet := false

	if val, set := getStringParamB("address"); set {
		server.Address = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getStringParamB("port"); set {
		server.Port = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getStringParamB("key"); set {
		server.TLSKey = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getStringParamB("cert"); set {
		server.TLSCert = val
		isAddrSet = isAddrSet || set
	}

	if val, set := getStringParamB("socket"); set {
		server.Socket = val
		isSocketSet = isSocketSet || set
	}

	if isAddrSet && isSocketSet {
		return nil, errors.New("--socket flag cannot be used with --address, --port, --key nor --cert")
	}

	// Do not use saved Socket if address was manually set.
	if isAddrSet && server.Socket != "" {
		server.Socket = ""
	}

	disableThumbnails := v.GetBool("disablethumbnails")
	server.EnableThumbnails = !disableThumbnails

	disablePreviewResize := v.GetBool("disablepreviewresize")
	server.ResizePreview = !disablePreviewResize

	disableTypeDetectionByHeader := v.GetBool("disabletypedetectionbyheader")
	server.TypeDetectionByHeader = !disableTypeDetectionByHeader

	disableExec := v.GetBool("disableexec")
	server.EnableExec = !disableExec

	if server.EnableExec {
		log.Println("WARNING: Command Runner feature enabled!")
		log.Println("WARNING: This feature has known security vulnerabilities and should not")
		log.Println("WARNING: you fully understand the risks involved. For more information")
		log.Println("WARNING: read https://github.com/filebrowser/filebrowser/issues/5199")
	}

	if val, set := getStringParamB("tokenexpirationtime"); set {
		server.TokenExpirationTime = val
	}

	return server, nil
}

func getStringParamB(key string) (string, bool) {
	return v.GetString(key), v.IsSet(key)
}

func setupLog(logMethod string) {
	switch logMethod {
	case "stdout":
		log.SetOutput(os.Stdout)
	case "stderr":
		log.SetOutput(os.Stderr)
	case "":
		log.SetOutput(io.Discard)
	default:
		log.SetOutput(&lumberjack.Logger{
			Filename:   logMethod,
			MaxSize:    100,
			MaxAge:     14,
			MaxBackups: 10,
		})
	}
}

func quickSetup(d pythonData) error {
	log.Println("Performing quick setup")

	set := &settings.Settings{
		Key:                   generateKey(),
		Signup:                false,
		HideLoginButton:       true,
		CreateUserDir:         false,
		MinimumPasswordLength: settings.DefaultMinimumPasswordLength,
		UserHomeBasePath:      settings.DefaultUsersHomeBasePath,
		Defaults: settings.UserDefaults{
			Scope:          ".",
			Locale:         "en",
			SingleClick:    false,
			AceEditorTheme: v.GetString("defaults.aceeditortheme"),
			Perm: users.Permissions{
				Admin:    false,
				Execute:  true,
				Create:   true,
				Rename:   true,
				Modify:   true,
				Delete:   true,
				Share:    true,
				Download: true,
			},
		},
		AuthMethod: "",
		Branding:   settings.Branding{},
		Tus: settings.Tus{
			ChunkSize:  settings.DefaultTusChunkSize,
			RetryCount: settings.DefaultTusRetryCount,
		},
		Commands: nil,
		Shell:    nil,
		Rules:    nil,
	}

	var err error
	if _, noauth := getStringParamB("noauth"); noauth {
		set.AuthMethod = auth.MethodNoAuth
		err = d.store.Auth.Save(&auth.NoAuth{})
	} else {
		set.AuthMethod = auth.MethodJSONAuth
		err = d.store.Auth.Save(&auth.JSONAuth{})
	}
	if err != nil {
		return err
	}

	err = d.store.Settings.Save(set)
	if err != nil {
		return err
	}

	ser := &settings.Server{
		BaseURL: v.GetString("baseurl"),
		Port:    v.GetString("port"),
		Log:     v.GetString("log"),
		TLSKey:  v.GetString("key"),
		TLSCert: v.GetString("cert"),
		Address: v.GetString("address"),
		Root:    v.GetString("root"),
	}

	err = d.store.Settings.SaveServer(ser)
	if err != nil {
		return err
	}

	username := v.GetString("username")
	password := v.GetString("password")

	if password == "" {
		var pwd string
		pwd, err = users.RandomPwd(set.MinimumPasswordLength)
		if err != nil {
			return err
		}

		log.Printf("User '%s' initialized with randomly generated password: %s\n", username, pwd)
		password, err = users.ValidateAndHashPwd(pwd, set.MinimumPasswordLength)
		if err != nil {
			return err
		}
	} else {
		log.Printf("User '%s' initialize wth user-provided password\n", username)
	}

	if username == "" || password == "" {
		log.Fatal("username and password cannot be empty during quick setup")
	}

	user := &users.User{
		Username:     username,
		Password:     password,
		LockPassword: false,
	}

	set.Defaults.Apply(user)
	user.Perm.Admin = true

	return d.store.Users.Save(user)
}

func initConfig() {
	if cfgFile == "" {
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}
		v.AddConfigPath(".")
		v.AddConfigPath(home)
		v.AddConfigPath("/etc/filebrowser/")
		v.SetConfigName(".filebrowser")
	} else {
		v.SetConfigFile(cfgFile)
	}

	v.SetEnvPrefix("FB")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var configParseError v.ConfigParseError
		if errors.As(err, &configParseError) {
			panic(err)
		}
		cfgFile = "No config file used"
	} else {
		cfgFile = "Using config file: " + v.ConfigFileUsed()
	}
}
