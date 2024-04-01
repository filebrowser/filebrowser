package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
)

func init() {
	rootCmd.AddCommand(usersCmd)
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Users management utility",
	Long:  `Users management utility.`,
	Args:  cobra.NoArgs,
}

func printUsers(usrs []*users.User) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUsername\tScope\tLocale\tV. Mode\tS.Click\tAdmin\tExecute\tCreate\tRename\tModify\tDelete\tShare\tDownload\tPwd Lock")

	for _, u := range usrs {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%t\t%t\t%t\t%t\t%t\t%t\t%t\t%t\t%t\t%t\t\n",
			u.ID,
			u.Username,
			u.Scope,
			u.Locale,
			u.ViewMode,
			u.SingleClick,
			u.Perm.Admin,
			u.Perm.Execute,
			u.Perm.Create,
			u.Perm.Rename,
			u.Perm.Modify,
			u.Perm.Delete,
			u.Perm.Share,
			u.Perm.Download,
			u.LockPassword,
		)
	}

	w.Flush()
}

func parseUsernameOrID(arg string) (username string, id uint) {
	id64, err := strconv.ParseUint(arg, 10, 64)
	if err != nil {
		return arg, 0
	}
	return "", uint(id64)
}

func addUserFlags(flags *pflag.FlagSet) {
	flags.Bool("perm.admin", false, "admin perm for users")
	flags.Bool("perm.execute", true, "execute perm for users")
	flags.Bool("perm.create", true, "create perm for users")
	flags.Bool("perm.rename", true, "rename perm for users")
	flags.Bool("perm.modify", true, "modify perm for users")
	flags.Bool("perm.delete", true, "delete perm for users")
	flags.Bool("perm.share", true, "share perm for users")
	flags.Bool("perm.download", true, "download perm for users")
	flags.String("sorting.by", "name", "sorting mode (name, size or modified)")
	flags.Bool("sorting.asc", false, "sorting by ascending order")
	flags.Bool("lockPassword", false, "lock password")
	flags.StringSlice("commands", nil, "a list of the commands a user can execute")
	flags.String("scope", ".", "scope for users")
	flags.String("locale", "en", "locale for users")
	flags.String("viewMode", string(users.ListViewMode), "view mode for users")
	flags.Bool("singleClick", false, "use single clicks only")
}

func getViewMode(flags *pflag.FlagSet) users.ViewMode {
	viewMode := users.ViewMode(mustGetString(flags, "viewMode"))
	if viewMode != users.ListViewMode && viewMode != users.MosaicViewMode {
		checkErr(errors.New("view mode must be \"" + string(users.ListViewMode) + "\" or \"" + string(users.MosaicViewMode) + "\""))
	}
	return viewMode
}

//nolint:gocyclo
func getUserDefaults(flags *pflag.FlagSet, defaults *settings.UserDefaults, all bool) {
	visit := func(flag *pflag.Flag) {
		switch flag.Name {
		case "scope":
			defaults.Scope = mustGetString(flags, flag.Name)
		case "locale":
			defaults.Locale = mustGetString(flags, flag.Name)
		case "viewMode":
			defaults.ViewMode = getViewMode(flags)
		case "singleClick":
			defaults.SingleClick = mustGetBool(flags, flag.Name)
		case "perm.admin":
			defaults.Perm.Admin = mustGetBool(flags, flag.Name)
		case "perm.execute":
			defaults.Perm.Execute = mustGetBool(flags, flag.Name)
		case "perm.create":
			defaults.Perm.Create = mustGetBool(flags, flag.Name)
		case "perm.rename":
			defaults.Perm.Rename = mustGetBool(flags, flag.Name)
		case "perm.modify":
			defaults.Perm.Modify = mustGetBool(flags, flag.Name)
		case "perm.delete":
			defaults.Perm.Delete = mustGetBool(flags, flag.Name)
		case "perm.share":
			defaults.Perm.Share = mustGetBool(flags, flag.Name)
		case "perm.download":
			defaults.Perm.Download = mustGetBool(flags, flag.Name)
		case "commands":
			commands, err := flags.GetStringSlice(flag.Name)
			checkErr(err)
			defaults.Commands = commands
		case "sorting.by":
			defaults.Sorting.By = mustGetString(flags, flag.Name)
		case "sorting.asc":
			defaults.Sorting.Asc = mustGetBool(flags, flag.Name)
		}
	}

	if all {
		flags.VisitAll(visit)
	} else {
		flags.Visit(visit)
	}
}
