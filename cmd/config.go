package cmd

import (
	"encoding/json"
	nerrors "errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/filebrowser/filebrowser/v2/auth"
	"github.com/filebrowser/filebrowser/v2/errors"
	"github.com/filebrowser/filebrowser/v2/settings"
)

func init() {
	rootCmd.AddCommand(configCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration management utility",
	Long:  `Configuration management utility.`,
	Args:  cobra.NoArgs,
}

func addConfigFlags(flags *pflag.FlagSet) {
	addServerFlags(flags)
	addUserFlags(flags)
	flags.BoolP("signup", "s", false, "allow users to signup")
	flags.Bool("hideLoginButton", false, "hide login button from public pages")
	flags.Bool("createUserDir", false, "generate user's home directory automatically")
	flags.Uint("minimumPasswordLength", settings.DefaultMinimumPasswordLength, "minimum password length for new users")
	flags.String("shell", "", "shell command to which other commands should be appended")

	flags.String("auth.method", string(auth.MethodJSONAuth), "authentication type")
	flags.String("auth.header", "", "HTTP header for auth.method=proxy")
	flags.String("auth.command", "", "command for auth.method=hook")

	flags.String("recaptcha.host", "https://www.google.com", "use another host for ReCAPTCHA. recaptcha.net might be useful in China")
	flags.String("recaptcha.key", "", "ReCaptcha site key")
	flags.String("recaptcha.secret", "", "ReCaptcha secret")

	flags.String("branding.name", "", "replace 'File Browser' by this name")
	flags.String("branding.theme", "", "set the theme")
	flags.String("branding.color", "", "set the theme color")

	flags.String("branding.files", "", "path to directory with images and custom styles")
	flags.Bool("branding.disableExternal", false, "disable external links such as GitHub links")
	flags.Bool("branding.disableUsedPercentage", false, "disable used disk percentage graph")

	// NB: these are string so they can be presented as octal in the help text
	// as that's the conventional representation for modes in Unix.
	flags.String("fileMode", fmt.Sprintf("%O", settings.DefaultFileMode), "mode bits that new files are created with")
	flags.String("dirMode", fmt.Sprintf("%O", settings.DefaultDirMode), "mode bits that new directories are created with")

	flags.Uint64("tus.chunkSize", settings.DefaultTusChunkSize, "the tus chunk size")
	flags.Uint16("tus.retryCount", settings.DefaultTusRetryCount, "the tus retry count")
}

func getAuthMethod(v *viper.Viper, defaults ...interface{}) (settings.AuthMethod, map[string]interface{}, error) {
	methodStr := v.GetString("auth.method")
	method := settings.AuthMethod(methodStr)

	var defaultAuther map[string]interface{}
	if len(defaults) > 0 {
		if hasAuth := defaults[0]; hasAuth != true {
			for _, arg := range defaults {
				switch def := arg.(type) {
				case *settings.Settings:
					method = def.AuthMethod
				case auth.Auther:
					ms, err := json.Marshal(def)
					if err != nil {
						return "", nil, err
					}
					err = json.Unmarshal(ms, &defaultAuther)
					if err != nil {
						return "", nil, err
					}
				}
			}
		}
	}

	return method, defaultAuther, nil
}

func getProxyAuth(v *viper.Viper, defaultAuther map[string]interface{}) (auth.Auther, error) {
	header := v.GetString("auth.header")
	if header == "" {
		header = defaultAuther["header"].(string)
	}

	if header == "" {
		return nil, nerrors.New("you must set the flag 'auth.header' for method 'proxy'")
	}

	return &auth.ProxyAuth{Header: header}, nil
}

func getNoAuth() auth.Auther {
	return &auth.NoAuth{}
}

func getJSONAuth(v *viper.Viper, defaultAuther map[string]interface{}) (auth.Auther, error) {
	jsonAuth := &auth.JSONAuth{}
	host := v.GetString("recaptcha.host")
	key := v.GetString("recaptcha.key")
	secret := v.GetString("recaptcha.secret")

	if key == "" {
		if kmap, ok := defaultAuther["recaptcha"].(map[string]interface{}); ok {
			key = kmap["key"].(string)
		}
	}

	if secret == "" {
		if smap, ok := defaultAuther["recaptcha"].(map[string]interface{}); ok {
			secret = smap["secret"].(string)
		}
	}

	if key != "" && secret != "" {
		jsonAuth.ReCaptcha = &auth.ReCaptcha{
			Host:   host,
			Key:    key,
			Secret: secret,
		}
	}
	return jsonAuth, nil
}

func getHookAuth(v *viper.Viper, defaultAuther map[string]interface{}) (auth.Auther, error) {
	command := v.GetString("auth.command")
	if command == "" {
		command = defaultAuther["command"].(string)
	}

	if command == "" {
		return nil, nerrors.New("you must set the flag 'auth.command' for method 'hook'")
	}

	return &auth.HookAuth{Command: command}, nil
}

func getAuthentication(v *viper.Viper, defaults ...interface{}) (settings.AuthMethod, auth.Auther, error) {
	method, defaultAuther, err := getAuthMethod(v, defaults...)
	if err != nil {
		return "", nil, err
	}

	var auther auth.Auther
	switch method {
	case auth.MethodProxyAuth:
		auther, err = getProxyAuth(v, defaultAuther)
	case auth.MethodNoAuth:
		auther = getNoAuth()
	case auth.MethodJSONAuth:
		auther, err = getJSONAuth(v, defaultAuther)
	case auth.MethodHookAuth:
		auther, err = getHookAuth(v, defaultAuther)
	default:
		return "", nil, errors.ErrInvalidAuthMethod
	}

	if err != nil {
		return "", nil, err
	}

	return method, auther, nil
}

func printSettings(ser *settings.Server, set *settings.Settings, auther auth.Auther) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Sign up:\t%t\n", set.Signup)
	fmt.Fprintf(w, "Hide Login Button:\t%t\n", set.HideLoginButton)
	fmt.Fprintf(w, "Create User Dir:\t%t\n", set.CreateUserDir)
	fmt.Fprintf(w, "Minimum Password Length:\t%d\n", set.MinimumPasswordLength)
	fmt.Fprintf(w, "Auth Method:\t%s\n", set.AuthMethod)
	fmt.Fprintf(w, "Shell:\t%s\t\n", strings.Join(set.Shell, " "))
	fmt.Fprintln(w, "\nBranding:")
	fmt.Fprintf(w, "\tName:\t%s\n", set.Branding.Name)
	fmt.Fprintf(w, "\tFiles override:\t%s\n", set.Branding.Files)
	fmt.Fprintf(w, "\tDisable external links:\t%t\n", set.Branding.DisableExternal)
	fmt.Fprintf(w, "\tDisable used disk percentage graph:\t%t\n", set.Branding.DisableUsedPercentage)
	fmt.Fprintf(w, "\tColor:\t%s\n", set.Branding.Color)
	fmt.Fprintf(w, "\tTheme:\t%s\n", set.Branding.Theme)
	fmt.Fprintln(w, "\nServer:")
	fmt.Fprintf(w, "\tLog:\t%s\n", ser.Log)
	fmt.Fprintf(w, "\tPort:\t%s\n", ser.Port)
	fmt.Fprintf(w, "\tBase URL:\t%s\n", ser.BaseURL)
	fmt.Fprintf(w, "\tRoot:\t%s\n", ser.Root)
	fmt.Fprintf(w, "\tSocket:\t%s\n", ser.Socket)
	fmt.Fprintf(w, "\tAddress:\t%s\n", ser.Address)
	fmt.Fprintf(w, "\tTLS Cert:\t%s\n", ser.TLSCert)
	fmt.Fprintf(w, "\tTLS Key:\t%s\n", ser.TLSKey)
	fmt.Fprintf(w, "\tExec Enabled:\t%t\n", ser.EnableExec)
	fmt.Fprintln(w, "\nTUS:")
	fmt.Fprintf(w, "\tChunk size:\t%d\n", set.Tus.ChunkSize)
	fmt.Fprintf(w, "\tRetry count:\t%d\n", set.Tus.RetryCount)
	fmt.Fprintln(w, "\nDefaults:")
	fmt.Fprintf(w, "\tScope:\t%s\n", set.Defaults.Scope)
	fmt.Fprintf(w, "\tHideDotfiles:\t%t\n", set.Defaults.HideDotfiles)
	fmt.Fprintf(w, "\tLocale:\t%s\n", set.Defaults.Locale)
	fmt.Fprintf(w, "\tView mode:\t%s\n", set.Defaults.ViewMode)
	fmt.Fprintf(w, "\tSingle Click:\t%t\n", set.Defaults.SingleClick)
	fmt.Fprintf(w, "\tFile Creation Mode:\t%O\n", set.FileMode)
	fmt.Fprintf(w, "\tDirectory Creation Mode:\t%O\n", set.DirMode)
	fmt.Fprintf(w, "\tCommands:\t%s\n", strings.Join(set.Defaults.Commands, " "))
	fmt.Fprintf(w, "\tAce editor syntax highlighting theme:\t%s\n", set.Defaults.AceEditorTheme)
	fmt.Fprintf(w, "\tSorting:\n")
	fmt.Fprintf(w, "\t\tBy:\t%s\n", set.Defaults.Sorting.By)
	fmt.Fprintf(w, "\t\tAsc:\t%t\n", set.Defaults.Sorting.Asc)
	fmt.Fprintf(w, "\tPermissions:\n")
	fmt.Fprintf(w, "\t\tAdmin:\t%t\n", set.Defaults.Perm.Admin)
	fmt.Fprintf(w, "\t\tExecute:\t%t\n", set.Defaults.Perm.Execute)
	fmt.Fprintf(w, "\t\tCreate:\t%t\n", set.Defaults.Perm.Create)
	fmt.Fprintf(w, "\t\tRename:\t%t\n", set.Defaults.Perm.Rename)
	fmt.Fprintf(w, "\t\tModify:\t%t\n", set.Defaults.Perm.Modify)
	fmt.Fprintf(w, "\t\tDelete:\t%t\n", set.Defaults.Perm.Delete)
	fmt.Fprintf(w, "\t\tShare:\t%t\n", set.Defaults.Perm.Share)
	fmt.Fprintf(w, "\t\tDownload:\t%t\n", set.Defaults.Perm.Download)
	w.Flush()

	b, err := json.MarshalIndent(auther, "", "  ")
	if err != nil {
		return err
	}
	fmt.Printf("\nAuther configuration (raw):\n\n%s\n\n", string(b))
	return nil
}
