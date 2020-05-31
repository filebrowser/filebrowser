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
	flags.String("shell", "", "shell command to which other commands should be appended")

	flags.String("auth.method", string(auth.MethodJSONAuth), "authentication type")
	flags.String("auth.header", "", "HTTP header for auth.method=proxy")

	flags.String("recaptcha.host", "https://www.google.com", "use another host for ReCAPTCHA. recaptcha.net might be useful in China")
	flags.String("recaptcha.key", "", "ReCaptcha site key")
	flags.String("recaptcha.secret", "", "ReCaptcha secret")

	flags.String("branding.name", "", "replace 'File Browser' by this name")
	flags.String("branding.files", "", "path to directory with images and custom styles")
	flags.Bool("branding.disableExternal", false, "disable external links such as GitHub links")
}

//nolint:gocyclo
func getAuthentication(flags *pflag.FlagSet, defaults ...interface{}) (settings.AuthMethod, auth.Auther) {
	method := settings.AuthMethod(mustGetString(flags, "auth.method"))

	var defaultAuther map[string]interface{}
	if len(defaults) > 0 {
		if hasAuth := defaults[0]; hasAuth != true {
			for _, arg := range defaults {
				switch def := arg.(type) {
				case *settings.Settings:
					method = def.AuthMethod
				case auth.Auther:
					ms, err := json.Marshal(def)
					checkErr(err)
					err = json.Unmarshal(ms, &defaultAuther)
					checkErr(err)
				}
			}
		}
	}

	var auther auth.Auther
	if method == auth.MethodProxyAuth {
		header := mustGetString(flags, "auth.header")

		if header == "" {
			header = defaultAuther["header"].(string)
		}

		if header == "" {
			checkErr(nerrors.New("you must set the flag 'auth.header' for method 'proxy'"))
		}

		auther = &auth.ProxyAuth{Header: header}
	}

	if method == auth.MethodNoAuth {
		auther = &auth.NoAuth{}
	}

	if method == auth.MethodJSONAuth {
		jsonAuth := &auth.JSONAuth{}
		host := mustGetString(flags, "recaptcha.host")
		key := mustGetString(flags, "recaptcha.key")
		secret := mustGetString(flags, "recaptcha.secret")

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
		auther = jsonAuth
	}

	if auther == nil {
		panic(errors.ErrInvalidAuthMethod)
	}

	return method, auther
}

func printSettings(ser *settings.Server, set *settings.Settings, auther auth.Auther) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintf(w, "Sign up:\t%t\n", set.Signup)
	fmt.Fprintf(w, "Create User Dir:\t%t\n", set.CreateUserDir)
	fmt.Fprintf(w, "Auth method:\t%s\n", set.AuthMethod)
	fmt.Fprintf(w, "Shell:\t%s\t\n", strings.Join(set.Shell, " "))
	fmt.Fprintln(w, "\nBranding:")
	fmt.Fprintf(w, "\tName:\t%s\n", set.Branding.Name)
	fmt.Fprintf(w, "\tFiles override:\t%s\n", set.Branding.Files)
	fmt.Fprintf(w, "\tDisable external links:\t%t\n", set.Branding.DisableExternal)
	fmt.Fprintln(w, "\nServer:")
	fmt.Fprintf(w, "\tLog:\t%s\n", ser.Log)
	fmt.Fprintf(w, "\tPort:\t%s\n", ser.Port)
	fmt.Fprintf(w, "\tBase URL:\t%s\n", ser.BaseURL)
	fmt.Fprintf(w, "\tRoot:\t%s\n", ser.Root)
	fmt.Fprintf(w, "\tSocket:\t%s\n", ser.Socket)
	fmt.Fprintf(w, "\tAddress:\t%s\n", ser.Address)
	fmt.Fprintf(w, "\tTLS Cert:\t%s\n", ser.TLSCert)
	fmt.Fprintf(w, "\tTLS Key:\t%s\n", ser.TLSKey)
	fmt.Fprintln(w, "\nDefaults:")
	fmt.Fprintf(w, "\tScope:\t%s\n", set.Defaults.Scope)
	fmt.Fprintf(w, "\tLocale:\t%s\n", set.Defaults.Locale)
	fmt.Fprintf(w, "\tView mode:\t%s\n", set.Defaults.ViewMode)
	fmt.Fprintf(w, "\tCommands:\t%s\n", strings.Join(set.Defaults.Commands, " "))
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
	checkErr(err)
	fmt.Printf("\nAuther configuration (raw):\n\n%s\n\n", string(b))
}
