package cmd

import (
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/filebrowser/filebrowser/v2/settings"
	"github.com/filebrowser/filebrowser/v2/users"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	rootCmd.AddCommand(usersCmd)
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Users management utility",
	Long:  `Users management utility.`,
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(0)
	},
}

func printUsers(users []*users.User) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tUsername\tScope\tLocale\tV. Mode\tAdmin\tExecute\tCreate\tRename\tModify\tDelete\tShare\tDownload\tPwd Lock")

	for _, user := range users {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%t\t%t\t%t\t%t\t%t\t%t\t%t\t%t\t%t\t\n",
			user.ID,
			user.Username,
			user.Scope,
			user.Locale,
			user.ViewMode,
			user.Perm.Admin,
			user.Perm.Execute,
			user.Perm.Create,
			user.Perm.Rename,
			user.Perm.Modify,
			user.Perm.Delete,
			user.Perm.Share,
			user.Perm.Download,
			user.LockPassword,
		)
	}

	w.Flush()
}

func usernameOrIDRequired(cmd *cobra.Command, args []string) error {
	username, _ := cmd.Flags().GetString("username")
	id, _ := cmd.Flags().GetUint("id")

	if username == "" && id == 0 {
		return errors.New("'username' of 'id' flag required")
	}

	return nil
}

func addUserFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("perm.admin", false, "admin perm for users")
	cmd.Flags().Bool("perm.execute", true, "execute perm for users")
	cmd.Flags().Bool("perm.create", true, "create perm for users")
	cmd.Flags().Bool("perm.rename", true, "rename perm for users")
	cmd.Flags().Bool("perm.modify", true, "modify perm for users")
	cmd.Flags().Bool("perm.delete", true, "delete perm for users")
	cmd.Flags().Bool("perm.share", true, "share perm for users")
	cmd.Flags().Bool("perm.download", true, "download perm for users")
	cmd.Flags().String("sorting.by", "name", "sorting mode (name, size or modified)")
	cmd.Flags().Bool("sorting.asc", false, "sorting by ascending order")
	cmd.Flags().Bool("lockPassword", false, "lock password")
	cmd.Flags().StringSlice("commands", nil, "a list of the commands a user can execute")
	cmd.Flags().String("scope", ".", "scope for users")
	cmd.Flags().String("locale", "en", "locale for users")
	cmd.Flags().String("viewMode", string(users.ListViewMode), "view mode for users")
}

func getViewMode(cmd *cobra.Command) users.ViewMode {
	viewMode := users.ViewMode(mustGetString(cmd, "viewMode"))
	if viewMode != users.ListViewMode && viewMode != users.MosaicViewMode {
		checkErr(errors.New("view mode must be \"" + string(users.ListViewMode) + "\" or \"" + string(users.MosaicViewMode) + "\""))
	}
	return viewMode
}

func getUserDefaults(cmd *cobra.Command, defaults *settings.UserDefaults, all bool) {
	visit := func(flag *pflag.Flag) {
		switch flag.Name {
		case "scope":
			defaults.Scope = mustGetString(cmd, flag.Name)
		case "locale":
			defaults.Locale = mustGetString(cmd, flag.Name)
		case "viewMode":
			defaults.ViewMode = getViewMode(cmd)
		case "perm.admin":
			defaults.Perm.Admin = mustGetBool(cmd, flag.Name)
		case "perm.execute":
			defaults.Perm.Execute = mustGetBool(cmd, flag.Name)
		case "perm.create":
			defaults.Perm.Create = mustGetBool(cmd, flag.Name)
		case "perm.rename":
			defaults.Perm.Rename = mustGetBool(cmd, flag.Name)
		case "perm.modify":
			defaults.Perm.Modify = mustGetBool(cmd, flag.Name)
		case "perm.delete":
			defaults.Perm.Delete = mustGetBool(cmd, flag.Name)
		case "perm.share":
			defaults.Perm.Share = mustGetBool(cmd, flag.Name)
		case "perm.download":
			defaults.Perm.Download = mustGetBool(cmd, flag.Name)
		case "commands":
			commands, err := cmd.Flags().GetStringSlice(flag.Name)
			checkErr(err)
			defaults.Commands = commands
		case "sorting.by":
			defaults.Sorting.By = mustGetString(cmd, flag.Name)
		case "sorting.asc":
			defaults.Sorting.Asc = mustGetBool(cmd, flag.Name)
		}
	}

	if all {
		cmd.Flags().VisitAll(visit)
	} else {
		cmd.Flags().Visit(visit)
	}
}
